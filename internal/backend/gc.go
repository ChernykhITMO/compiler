package backend

import (
	"fmt"

	"github.com/ChernykhITMO/compiler/internal/bytecode"
)

func (vm *VM) newObject(t bytecode.ObjectType) *bytecode.Object {
	if vm.heap.MaxObjects == 0 {
		vm.heap.MaxObjects = 8
	}

	if vm.heap.NumObjects >= vm.heap.MaxObjects {
		vm.gc()
	}

	obj := &bytecode.Object{
		Type: t,
		Next: vm.heap.Head,
	}
	vm.heap.Head = obj
	vm.heap.NumObjects++

	fmt.Printf("[GC] newObject type=%v -> NumObjects=%d\n", t, vm.heap.NumObjects)

	return obj
}

func (vm *VM) gc() {
	before := vm.heap.NumObjects
	fmt.Printf("[GC] start: NumObjects=%d, MaxObjects=%d\n",
		before, vm.heap.MaxObjects)

	vm.markRoots()
	vm.sweep()

	after := vm.heap.NumObjects
	fmt.Printf("[GC] end:   NumObjects=%d, MaxObjects=%d (collected %d)\n",
		after, vm.heap.MaxObjects, before-after)

	if vm.heap.NumObjects < 8 {
		vm.heap.MaxObjects = 8
	} else {
		vm.heap.MaxObjects = vm.heap.NumObjects * 2
	}
}

func (vm *VM) markRoots() {
	for _, r := range vm.roots {
		if r.locals != nil {
			for i := range *r.locals {
				vm.markValue(&(*r.locals)[i])
			}
		}
		if r.stack != nil {
			for i := range *r.stack {
				vm.markValue(&(*r.stack)[i])
			}
		}
	}
}

func (vm *VM) markValue(v *bytecode.Value) {
	if v == nil || v.Obj == nil {
		return
	}
	vm.markObject(v.Obj)
}

func (vm *VM) markObject(o *bytecode.Object) {
	if o == nil || o.Mark {
		return
	}
	o.Mark = true

	switch o.Type {
	case bytecode.ObjArray:
		for i := range o.Items {
			vm.markValue(&o.Items[i])
		}
	}
}

func (vm *VM) sweep() {
	var prev *bytecode.Object
	curr := vm.heap.Head

	for curr != nil {
		if !curr.Mark {
			unreached := curr

			if prev != nil {
				prev.Next = curr.Next
			} else {
				vm.heap.Head = curr.Next
			}

			curr = curr.Next
			vm.heap.NumObjects--

			*unreached = bytecode.Object{}
		} else {
			curr.Mark = false
			prev = curr
			curr = curr.Next
		}
	}
}
