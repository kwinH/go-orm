package oorm

import (
	"github.com/kwinH/go-oorm/schema"
	"reflect"
	"strings"
	"sync"
)

type WithFunc func(*DB)

type With struct {
	*schema.With
	Callback WithFunc
}

func (d *DB) With(name string, callbacks ...WithFunc) *DB {
	db := d.getInstance()

	if db.withs == nil {
		db.withs = make(map[string]WithFunc)
	}

	if db.childWiths == nil {
		db.childWiths = make(map[string][]WithFunc)
	}

	var callbackArr = make([]WithFunc, 0)
	callbacksLen := len(callbacks)
	names := strings.Split(name, ".")

	for i, _ := range names {
		callback := func(db *DB) {}
		if callbacksLen > i {
			if callbacks[i] != nil {
				callback = callbacks[i]
			}
		}
		callbackArr = append(callbackArr, callback)
	}

	name = names[0]
	db.withs[name] = callbackArr[0]

	names = names[1:]
	callbackArr = callbackArr[1:]
	namesStr := strings.Join(names, ".")

	db.childWiths[namesStr] = callbackArr

	return db
}

func (d *DB) makeWiths(tableInfo *schema.Schema) []*With {
	withs := make([]*With, 0)
	for key, callback := range d.withs {
		if w, ok := tableInfo.Withs[key]; ok {
			with := &With{
				With:     w,
				Callback: callback,
			}

			withs = append(withs, with)
		}
	}
	return withs
}

func (d *DB) getWiths(withs []*With, dest reflect.Value) {
	for _, with := range withs {

		var localKeyValue interface{}

		val := dest.FieldByName(with.LocalKey.Name)

		if !val.IsValid() || val.IsZero() {
			continue
		}

		localKeyValue = val.Interface()

		with.Values = append(with.Values, localKeyValue)
	}
}

func (d *DB) relationships(withs []*With) {
	wg := &sync.WaitGroup{}
	for _, with := range withs {
		wg.Add(1)
		go d.setWithRelationships(with, wg)
	}

	wg.Wait()
}

func (d *DB) setWithRelationships(with *With, wg *sync.WaitGroup) {
	defer wg.Done()

	if with.Values == nil {
		return
	}

	joinResults := schema.MakeSlice(with.ModelType).Elem()

	db := d.ClonePure(1)

	for modelName, funcList := range d.childWiths {
		db.With(modelName, funcList...)
	}

	with.Callback(db)
	err := db.Where(with.ForeignKey.FieldName, "in", with.Values).Get(joinResults.Addr().Interface())

	if err != nil {
		d.AddError(err)
		return
	}

	for i := 0; i < joinResults.Len(); i++ {
		val := joinResults.Index(i)
		key := val.FieldByName(with.ForeignKey.Name).Interface()
		with.Relationships[key] = append(with.Relationships[key], val)
	}
}

func (d *DB) setDestRelationships(dests []reflect.Value, withs []*With, value reflect.Value) {
	for _, dest := range dests {
		wg := &sync.WaitGroup{}
		for _, with := range withs {
			wg.Add(1)
			d.setDestRelationship(with, dest, wg)
		}
		wg.Wait()
		if value.Kind() == reflect.Slice {
			value.Set(reflect.Append(value, dest))
		} else {
			value.Set(dest)
			break
		}
	}
}

func (d *DB) setDestRelationship(with *With, dest reflect.Value, wg *sync.WaitGroup) {
	defer wg.Done()
	relationshipValues := with.Relationships[dest.FieldByName(with.LocalKey.Name).Interface()]
	if relationshipValues != nil {
		switch with.Type {
		case schema.One:
			dest.FieldByName(with.Name).Set(relationshipValues[0])
		case schema.Many:
			joinResults := schema.MakeSlice(with.ModelType).Elem()
			dest.FieldByName(with.Name).Set(reflect.Append(joinResults, relationshipValues...))
		}
	}
}
