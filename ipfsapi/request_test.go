package ipfsapi

import (
	"fmt"
	"github.com/cheekybits/is"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var host = "http://localhost:5001"
var examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"

func TestAdd(t *testing.T) {
	f, err := os.Open("../main.go")
	if err != nil {
		t.Error(err)
		return
	}
	c := NewClient(host)
	obj, err := c.Add(f, f.Name())
	fmt.Println(obj, err)
}

func TestCat(t *testing.T) {
	r, _ := NewClient(host).Cat(fmt.Sprintf("/ipfs/%s/readme", examplesHash))
	b, _ := ioutil.ReadAll(r)
	fmt.Println(string(b))
}

func TestClient_AddDir(t *testing.T) {
	r, _ := NewClient(host).AddDir("../ipfsapi")
	fmt.Println(r)
}

func TestClient_AddLink(t *testing.T) {
	r, _ := NewClient(host).AddLink("./client.go")
	fmt.Println(r)
}

func TestClient_ID(t *testing.T) {
	r, _ := NewClient(host).ID()
	fmt.Println(r)
}

func TestClient_List(t *testing.T) {
	r, err := NewClient(host).List("/ipfs/" + examplesHash)
	if err != nil {
		t.Error(err)
		return
	}
	for _, v := range r {
		fmt.Println(v)
	}
}

func TestClient_Pins(t *testing.T) {
	r, err := NewClient(host).Pins()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r)
}

func TestClient_FindPeer(t *testing.T) {
	r, err := NewClient(host).FindPeer("QmYt7wv5EEarni9r3VxoTrhqw3ULQh42MuWb6c1bhaJrnT")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r)
}

func TestClient_Refs(t *testing.T) {
	r, err := NewClient(host).Refs("QmYt7wv5EEarni9r3VxoTrhqw3ULQh42MuWb6c1bhaJrnT", true)
	if err != nil {
		t.Error(err)
		return
	}
	for i := range r {
		fmt.Println(i)
	}
}

func TestClient_Patch_RmLink(t *testing.T) {
	is := is.New(t)
	s := NewClient(host)
	newRoot, err := s.PatchRmLink(examplesHash, "rm-link", "about")
	is.Nil(err)
	is.Equal(newRoot.Hash, "QmPmCJpciopaZnKcwymfQyRAEjXReR6UL2rdSfEscZfzcp")
}

func TestPatchLink(t *testing.T) {
	is := is.New(t)
	s := NewClient(host)

	newRoot, err := s.PatchAddLink(examplesHash, "about", "QmUXTtySmd7LD4p6RG6rZW6RuUuPZXTtNMmRQ6DSQo3aMw", true)
	is.Nil(err)
	is.Equal(newRoot.Hash, "QmVfe7gesXf4t9JzWePqqib8QSifC1ypRBGeJHitSnF7fA")
}

func TestClient_Get(t *testing.T) {
	err := NewClient(host).Get(examplesHash, "./test")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestClient_ResolvePath(t *testing.T) {
	r, err := NewClient(host).ResolvePath(examplesHash)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r.CumulativeSize)
}

func TestClient_Version(t *testing.T) {
	r, err := NewClient(host).Version()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r)
}

func TestClient_DagPut(t *testing.T) {
	r, err := NewClient(host).DagPut(`{"x": "abc","y":"def"}`, "json", "cbor")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r)
}

type JTSData struct {
	X	string	`json:"x"`
	Y	string	`json:"y"`
}



func TestClient_DagGet(t *testing.T) {
	v := JTSData{}
	err := NewClient(host).DagGet("zdpuAt47YjE9XTgSxUBkiYCbmnktKajQNheQBGASHj3FfYf8M", &v)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(v)
}

func TestPubSub(t *testing.T) {
	is := is.New(t)
	s := NewClient(host)

	var (
		topic = "test"
		sub *PubSubSubscription
		err error
	)

	t.Log("subscribing...")
	sub, err = s.PubSubSubscribe(topic)
	is.Nil(err)
	is.NotNil(sub)
	t.Log("sub: done")

	m, err := sub.Next()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(m.Data()))
	m, err = sub.Next()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(m.Data()))
	m, err = sub.Next()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(m.Data()))

	time.Sleep(10 * time.Millisecond)

	t.Log("publishing...")
	is.Nil(s.PubSubPublish(topic, "Hello World!"))
	t.Log("pub: done")
	time.Sleep(2 * time.Second)
	t.Log("next()...")
	r, err := sub.Next()
	t.Log("next: done. ")

	is.Nil(err)
	is.NotNil(r)
	is.Equal(r.Data(), "Hello World!")

	sub2, err := s.PubSubSubscribe(topic)
	is.Nil(err)
	is.NotNil(sub2)

	is.Nil(s.PubSubPublish(topic, "Hallo Welt!"))

	r, err = sub2.Next()
	is.Nil(err)
	is.NotNil(r)
	is.Equal(r.Data(), "Hallo Welt!")

	r, err = sub.Next()
	is.NotNil(r)
	is.Nil(err)
	is.Equal(r.Data(), "Hallo Welt!")

	is.Nil(sub.Cancel())
}

func TestShell_FileList(t *testing.T) {
	r, err := NewClient(host).FileList(examplesHash)
	if err != nil {
		t.Error(err)
		return
	}
	for _, v := range r.Links {
		fmt.Println(v)
	}
	fmt.Println(r.Type, r.Size, r.Hash)
}