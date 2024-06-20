package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/stsg/gophkeeper2/pkg/logger"
	restype "github.com/stsg/gophkeeper2/pkg/model"
	"github.com/stsg/gophkeeper2/pkg/model/enum"
	"github.com/stsg/gophkeeper2/server/model"
	"github.com/stsg/gophkeeper2/server/model/errs"
)

//go:generate mockgen -source=resource_repository.go -destination=../mocks/repositories/resource_repository.go -package=repositories

type ResourceRepository interface {
	Save(ctx context.Context, resource *model.Resource) error
	Update(ctx context.Context, resource *model.Resource) error
	Get(ctx context.Context, resId int32, userId int32) (*model.Resource, error)
	GetResDescriptionsByType(ctx context.Context, userId int32, resType enum.ResourceType) ([]*model.ResourceDescription, error)
	Delete(ctx context.Context, resId int32, userId int32) error
}

type resourceRepository struct {
	log *zap.SugaredLogger
	db  DBProvider
}

func NewResourceRepository(db DBProvider) ResourceRepository {
	return &resourceRepository{log: logger.NewLogger("res-repo"), db: db}
}

func (r *resourceRepository) Save(ctx context.Context, resource *model.Resource) error {
	r.log.Infof("Saving resource: %v", resource)
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return errs.DbError{Err: err}
	}
	defer conn.Release()
	var resId int32
	row := conn.QueryRow(
		ctx,
		"insert into resources(user_id, type, data, meta) values ($1, $2, $3, $4) RETURNING id",
		resource.UserId,
		resource.Type,
		resource.Data,
		resource.Meta,
	)
	err = row.Scan(&resId)
	if err != nil {
		r.log.Errorf("failed to scan resId: %v", err)
		return errs.DbError{Err: err}
	}
	resource.Id = resId
	r.log.Infof("Resource saved: %v", resource)
	return nil
}

func (r *resourceRepository) Update(ctx context.Context, resource *model.Resource) error {
	r.log.Infof("Updating resource: %v", resource)
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return errs.DbError{Err: err}
	}
	defer conn.Release()
	var resId int32
	row := conn.QueryRow(
		ctx,
		"insert into resources(id, user_id, type, data, meta) values ($1, $2, $3, $4, $5) "+
			"ON CONFLICT (id) DO UPDATE SET data = excluded.data, meta = excluded.meta "+
			"RETURNING id",
		resource.Id,
		resource.UserId,
		resource.Type,
		resource.Data,
		resource.Meta,
	)
	err = row.Scan(&resId)
	if err != nil {
		r.log.Errorf("failed to scan resId: %v", err)
		return errs.DbError{Err: err}
	}
	resource.Id = resId
	r.log.Infof("Resource updated: %v", resource)
	return nil
}

func (r *resourceRepository) Get(ctx context.Context, resId int32, userId int32) (*model.Resource, error) {
	r.log.Infof("Getting '%d' resource of '%d' user", resId, userId)
	var result model.Resource
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return nil, errs.DbError{Err: err}
	}
	defer conn.Release()
	row := conn.QueryRow(ctx, "select id, user_id, type, meta, data from resources where id = $1 and user_id = $2", resId, userId)
	err = row.Scan(&result.Id, &result.UserId, &result.Type, &result.Meta, &result.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		r.log.Warnf("There is no '%d' resource of '%d' user", resId, userId)
		return nil, errs.ErrResNotFound
	}
	if err != nil {
		r.log.Errorf("failed to parse scan resourse '%d' result row: %v", resId, err)
		return nil, errs.DbError{Err: err}
	}
	return &result, nil
}

func (r *resourceRepository) GetResDescriptionsByType(
	ctx context.Context,
	userId int32,
	resType enum.ResourceType,
) ([]*model.ResourceDescription, error) {
	r.log.Infof("Getting descriptions of '%d' user's resourses by type %s", userId, restype.TypeToArg[resType])
	var results []*model.ResourceDescription
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return results, errs.DbError{Err: err}
	}
	defer conn.Release()

	var rows pgx.Rows
	if resType == enum.Nan {
		r.log.Infof("Getting all resource descriptions of '%d' user", userId)
		rows, err = conn.Query(
			ctx,
			"select id, meta, type from resources where user_id = $1",
			userId,
		)
	} else {
		r.log.Infof("Getting '%s' resource descriptions of '%d' user", restype.TypeToArg[resType], userId)
		rows, err = conn.Query(
			ctx,
			"select id, meta, type from resources where user_id = $1 and type = $2",
			userId,
			resType,
		)
	}
	if err != nil {
		r.log.Errorf("failed to query resources for '%d' user: %v", userId, err)
		return nil, errs.DbError{Err: err}
	}
	defer rows.Close()
	for rows.Next() {
		resDescr := &model.ResourceDescription{}
		err := rows.Scan(&resDescr.Id, &resDescr.Meta, &resDescr.Type)
		if err != nil {
			r.log.Errorf("failed to scan '%s' resources of userId '%d': %v", restype.TypeToArg[resType], userId, err)
			return nil, errs.DbError{Err: fmt.Errorf("failed to read '%d' resources of userId '%d': %v", resType, userId, err)}
		}
		results = append(results, resDescr)
	}
	return results, err
}

func (r *resourceRepository) Delete(ctx context.Context, resId int32, userId int32) error {
	r.log.Infof("Deletting '%d' resource of '%d' user", resId, userId)
	conn, err := r.db.GetConnection(ctx)
	if err != nil {
		r.log.Errorf("failed to get db connection: %v", err)
		return errs.DbError{Err: err}
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "delete from resources where id = $1 and user_id = $2", resId, userId)
	if err != nil {
		r.log.Errorf("failed to delete  '%d' resource of '%d' user: %v", resId, userId, err)
		return errs.DbError{Err: err}
	}
	return nil
}
