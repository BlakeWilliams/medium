package medium

type (
	RouteBase[T Action] struct {
		router *Router[T]
	}
)

func (r *RouteBase[T]) Get(path string, handler HandlerFunc[T]) {
	r.router.Get(path, handler)
}
