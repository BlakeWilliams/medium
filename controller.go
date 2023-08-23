package medium

import (
	"reflect"
)

type Controller[RequestData any, ControllerData any] struct {
	// Hack to instantiate via reflect
	_t ControllerData
}

type ControllerHandler[RequestData any] interface {
	Handle(ResponseWriter, *Request[RequestData])
}

type ControllerHandlerFunc[RequestData any] func(ResponseWriter, *Request[RequestData])

func (c ControllerHandlerFunc[RequestData]) Handle(w ResponseWriter, r *Request[RequestData]) {
	c(w, r)
}

type Beforable[RequestData, ControllerData any] interface {
	BeforeHandler(w ResponseWriter, r *Request[RequestData], data ControllerData) bool
}

func (c *Controller[RequestData, ControllerData]) Get(
	path string,
	handler func(ResponseWriter, *Request[RequestData], ControllerData),
) ControllerHandler[RequestData] {
	return ControllerHandlerFunc[RequestData](func(w ResponseWriter, r *Request[RequestData]) {
		var ci any = c
		data := reflect.New(reflect.TypeOf(c._t)).Interface().(ControllerData)

		if beforable, ok := ci.(Beforable[RequestData, ControllerData]); ok {
			ok := beforable.BeforeHandler(w, r, data)

			if !ok {
				return
			}
		}

		handler(w, r, data)
	})
}

// type controllerRegisterable

type ControllerFunc[RouterData any, ControllerData any] func(ResponseWriter, *Request[RouterData], ControllerData)

// type BasicRouter[RouterData any, ControllerData any] struct {
// 	routable Routable[RouterData]
// 	_t       ControllerData
// }
