package service

import (
	"context"
	"tdg/internal/cache"
	"tdg/internal/repository"
)

type BaseService struct {
	Ctx   context.Context
	Debug bool
}

type AllRepository struct {
	ITrueRepository repository.ITrueRepository
	ICacheClient    cache.ICacheClient
}

type AllService struct {
	ITrueService ITrueService
}
