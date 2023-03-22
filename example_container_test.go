package container

type UserManager struct {
	Cache1 Cache
	Cache2 Cache `inject:"tag:redis"`
	Repo   *UserRepo
}

type Cache interface {
	Name() string
}

type FileCache struct{}

func (r *FileCache) Name() string {
	return "file"
}

type RedisCache struct{}

func (r *RedisCache) Name() string {
	return "redis"
}

type UserRepo struct{}

func Example() {
	fileCache := Query[*FileCache](None)
	redisCache := Query[*RedisCache](None)
	Provide[Cache](fileCache, None)
	Provide[Cache](redisCache, "redis")

	userManager := Query[*UserManager](None)
	if userManager.Cache1.Name() != "file" {
		panic("expected file, got " + userManager.Cache1.Name())
	}
	if userManager.Cache2.Name() != "redis" {
		panic("expected redis, got " + userManager.Cache2.Name())
	}
}
