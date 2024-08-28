package orm

import (
	"fmt"
	"testing"
)

func TestHasOne(t *testing.T) {
	type Contact struct {
		Model
		UserId uint
		Mobile string
		Email  string
	}

	type User struct {
		Model
		UserName string
		Password string
		Nickname string
		Status   string
		Avatar   string
		Contact  Contact
	}

	var user = &User{}

	err := orm.With("Contact").Find(user, 1)

	if err != nil {
		t.Error(err)
		return
	}
	t.Log("user.Contact.Mobile", user.Contact.Mobile)
}

func TestBelongsTo(t *testing.T) {
	type User struct {
		Model
		UserName string
		Password string
		Nickname string
		Status   string
		Avatar   string
	}

	type Contact struct {
		Model
		UserId uint
		Mobile string
		Email  string
		User   User `db:"localKey:UserId;foreignKey:Id"`
	}

	var contact = &Contact{}

	err := orm.With("User").Find(contact, 2)

	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(contact.UserId)
}

func TestHasOneCreate(t *testing.T) {
	type Contact struct {
		Model
		UserId uint
		Mobile string
	}

	type User struct {
		Model
		UserName string
		Contact  []Contact
		Contact1 []Contact
	}

	var user = &User{
		UserName: "kwinwong",
		Contact: []Contact{{
			Mobile: "13758665977",
		}, {
			Mobile: "13589217699",
		},
		},
		Contact1: []Contact{{
			Mobile: "13758665977",
		}, {
			Mobile: "13589217699",
		},
		},
	}

	res, err := orm.With("Contact").With("Contact1").Create(user)

	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("res %#v\n", res)

	fmt.Printf("user.Id %v\n", user.Id)
	for _, contact := range user.Contact {
		fmt.Printf("contact.Id %v\n", contact.Id)
	}

	t.Log(user.Id)
}

func TestHasOneUpdate(t *testing.T) {
	type Contact struct {
		Model
		UserId uint
		Mobile string
	}

	type User struct {
		Model
		UserName string
		Contact  Contact
	}

	user := &User{}
	err := orm.With("Contact").Find(user, 2)

	if err != nil {
		t.Error(err)
	}

	user.Contact.Mobile = "1389650755"

	res, err := orm.With("Contact").Update(user)

	fmt.Printf("%#v %#v\n", res, err)
}
