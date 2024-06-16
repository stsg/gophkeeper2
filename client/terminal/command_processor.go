package terminal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/stsg/gophkeeper2/client/model/resources"
	"github.com/stsg/gophkeeper2/client/services"
	"github.com/stsg/gophkeeper2/pkg/model"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/pkg/shutdown"
)

const (
	maxCapacity   = 1024 * 1024
	successResult = "success"
	helpMsg       = "" +
		"\n" +
		"available commands:\n" +
		"\n" +
		"	'exit' - exit client\n" +
		"\n" +
		"	'clear' - to clear terminal\n" +
		"\n" +
		"	'login' - to login\n" +
		"	'register' - to register\n" +
		"\n" +
		"	's [type]' - save resource, where 'type' is: lp - LoginPassword, fl - File, bc - BankCard\n" +
		"\n" +
		"	'u [id]' - update resource\n" +
		"	'd [id]' - delete resource by id\n" +
		"	'l [type]' - get resources by type, where 'type' is: lp - LoginPassword, fl - File, bc - BankCard\n	or get all if type is empty\n" +
		"	'g [id]' - get loginPassword or BankCard by id\n" +
		"	'gf [id]' - get file by id\n"
)

type CommandParser interface {
	Start(exit chan struct{})
	InitScanner()
}

type commandParser struct {
	scanner         *bufio.Scanner
	authService     services.AuthService
	resourceService services.ResourceService
	exitHandler     shutdown.ExitHandler
	commands        map[string]func(args []string) (string, error)
	quit            chan struct{}
}

func NewCommandParser(
	buildVersion string,
	buildDate string,
	authService services.AuthService,
	resourceService services.ResourceService,
	eh shutdown.ExitHandler,
	quit chan struct{},
) CommandParser {
	fmt.Printf("buildVersion='%s' buildDate='%s'\n%s\n", buildVersion, buildDate, helpMsg)
	cp := &commandParser{
		authService:     authService,
		resourceService: resourceService,
		exitHandler:     eh,
		quit:            quit,
	}
	cp.commands = map[string]func(args []string) (string, error){
		"login":    cp.handleLogin,
		"register": cp.handleRegistration,
		"s":        cp.handleSave,
		"u":        cp.handleUpdate,
		"d":        cp.handleDelete,
		"l":        cp.handleList,
		"g":        cp.handleGet,
		"gf":       cp.handleGetFile,
		"clear":    cp.handleClear,
		"help":     cp.handleHelp,
		"exit":     cp.handleExit,
	}
	return cp
}

func (cp *commandParser) Start(exit chan struct{}) {
	commandHandled := make(chan struct{})
	for {
		go cp.processCommands(commandHandled)
		select {
		case <-commandHandled:
		case <-exit:
			fmt.Println("exit")
			return
		}
	}
}

func (cp *commandParser) processCommands(commandHandled chan struct{}) {
	defer func() {
		commandHandled <- struct{}{}
	}()
	cmd := cp.readString("")
	if len(cmd) == 0 {
		return
	}
	result, err := cp.handle(cmd)
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return
	}
	fmt.Printf("%s\n", result)
}

func (cp *commandParser) handle(input string) (string, error) {
	arr := strings.Split(input, " ")

	command := arr[0]
	args := arr[1:]

	if f, ok := cp.commands[command]; ok {
		return f(args)
	}
	return "", fmt.Errorf("command '%s' is not supported, type 'help' to display available commands", command)

}

func (cp *commandParser) handleClear(_ []string) (string, error) {
	fmt.Print("\033[H\033[2J")
	return "", nil
}

func (cp *commandParser) handleHelp(_ []string) (string, error) {
	fmt.Print(helpMsg)
	return "", nil
}

func (cp *commandParser) handleExit(_ []string) (string, error) {
	os.Exit(0)
	return "", nil
}

func (cp *commandParser) handleLogin(_ []string) (string, error) {
	login := cp.readString("input username")
	password := cp.readPassword()
	_, err := cp.authService.Login(context.Background(), login, password)
	return successResult, err
}

func (cp *commandParser) handleRegistration(_ []string) (string, error) {
	login := cp.readString("input username")
	password := cp.readPassword()
	_, err := cp.authService.Register(context.Background(), login, password)
	return successResult, err
}

func (cp *commandParser) handleGetFile(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("arg '[id]' is empty, type 'help' to display available commands format")
	}
	resId, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", err
	}
	cp.exitHandler.AddFuncInProcessing("getting file")
	defer cp.exitHandler.FuncFinished("getting file")
	path, err := cp.resourceService.GetFile(context.Background(), int32(resId))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("recieved file saved to: %v", path), nil
}

func (cp *commandParser) handleGet(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("arg '[id]' is empty, type 'help' to display available commands format")
	}
	resId, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", err
	}
	resDescription, err := cp.resourceService.Get(context.Background(), int32(resId))
	if err != nil {
		return "", err
	}

	return resDescription.Format(), nil
}

func (cp *commandParser) handleList(args []string) (string, error) {
	resType := enum.Nan
	if len(args) != 0 {
		if rType, ok := model.ArgToType[args[0]]; ok {
			resType = rType
		}
	}

	resDescriptions, err := cp.resourceService.GetDescriptions(context.Background(), resType)
	if err != nil {
		return "", err
	}
	var writer strings.Builder
	if len(resDescriptions) == 0 {
		_, err := writer.WriteString("empty")
		if err != nil {
			return "", err
		}
	}
	for _, resDescription := range resDescriptions {
		_, err := writer.WriteString(fmt.Sprintf("id: %d - type: '%s', descr: '%s'\n", resDescription.Id, model.TypeToArg[resDescription.Type], string(resDescription.Meta)))
		if err != nil {
			return "", err
		}
	}
	return writer.String(), nil
}

func (cp *commandParser) handleSave(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("arg '[type]' is empty, type 'help' to display available commands format")
	}
	var resource any
	var meta string
	resType := args[0]
	switch resType {
	case model.LoginPasswordArg:
		resource, meta = cp.readLoginPassword()
		return cp.saveTextResource(resource, meta, enum.LoginPassword)
	case model.BankCardArg:
		resource, meta = cp.readBankCard()
		return cp.saveTextResource(resource, meta, enum.BankCard)
	case model.FileArg:
		return cp.saveFile()
	default:
		return "", fmt.Errorf("resource type argument '%s' is not supported, type 'help' to display available types", resType)
	}
}

func (cp *commandParser) handleUpdate(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("arg '[id]' is empty, type 'help' to display available commands format")
	}
	resId, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", err
	}
	id := int32(resId)
	var resource any
	resDescription, err := cp.resourceService.Get(context.Background(), id)
	if err != nil {
		return "", err
	}
	var meta string
	switch resDescription.Resource.Type() {
	case enum.LoginPassword:
		resource, meta = cp.readLoginPassword()
		return cp.updateTextResource(id, resource, meta, enum.LoginPassword)
	case enum.BankCard:
		resource, meta = cp.readBankCard()
		return cp.updateTextResource(id, resource, meta, enum.BankCard)
	case enum.File:
		return "", fmt.Errorf("file update is not implemented, create a new")
	default:
		return "", fmt.Errorf("resource type argument '%d' is not supported, type 'help' to display available types", resDescription.Resource.Type())
	}
}

func (cp *commandParser) handleDelete(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("arg '[id]' is empty, type 'help' to display available commands format")
	}
	resId, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", err
	}
	err = cp.resourceService.Delete(context.Background(), int32(resId))
	if err != nil {
		return "", err
	}
	return "deleted", nil
}

func (cp *commandParser) saveTextResource(resource any, meta string, resType enum.ResourceType) (string, error) {
	resourceJson, err := json.Marshal(resource)
	if err != nil {
		return "", err
	}
	id, err := cp.resourceService.Save(context.Background(), resType, resourceJson, []byte(meta))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("saved successfully, id: %v", id), nil
}

func (cp *commandParser) updateTextResource(resId int32, resource any, meta string, resType enum.ResourceType) (string, error) {
	resourceJson, err := json.Marshal(resource)
	if err != nil {
		return "", err
	}
	err = cp.resourceService.Update(context.Background(), resId, resType, resourceJson, []byte(meta))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("updated successfully, id: %v", resId), nil
}

func (cp *commandParser) saveFile() (string, error) {
	filePath := cp.readString("input file path")
	meta := cp.readString("input description")
	cp.exitHandler.AddFuncInProcessing("sending file")
	defer cp.exitHandler.FuncFinished("sending file")
	id, err := cp.resourceService.SaveFile(context.Background(), filePath, []byte(meta))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

func (cp *commandParser) readLoginPassword() (*resources.LoginPassword, string) {
	login := cp.readString("input login")
	password := cp.readPassword()
	description := cp.readString("input description")

	return resources.NewLoginPassword(login, password), description
}

func (cp *commandParser) readBankCard() (*resources.BankCard, string) {
	number := cp.readString("input number")
	expireAt := cp.readString("input expireAt in format: MM/YY")
	name := cp.readString("input name")
	surname := cp.readString("input surname")
	description := cp.readString("input description")

	return resources.NewBankCard(number, expireAt, name, surname), description
}

func (cp *commandParser) readString(label string) string {
	if len(label) != 0 {
		fmt.Println(label)
	}
	fmt.Print(">>> ")
	if cp.scanner.Scan() {
		return cp.scanner.Text()
	}
	return "exit"
}

func (cp *commandParser) readPassword() string {
	fmt.Println("password:")
	fmt.Print(">>> ")
	// type conversion for compiling in windows platform
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()
	return string(bytePassword)
}

func (cp *commandParser) InitScanner() {
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(buf, maxCapacity)
	cp.scanner = scanner
}
