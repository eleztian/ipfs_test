package ipfsapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

type IPFSCommand int

const (
	ADD IPFSCommand = iota
	GET
)

var ipfscommands = []string{
	"add",
	"get",
	"swap/peers",
}

const (
	IPFS_API_URL    = "http://127.0.0.1:5001/api"
	IPFS_VERSION_V0 = "v0"
)

func (ic IPFSCommand) String() string {
	if len(ipfscommands) <= int(ic) {
		return ""
	}
	return ipfscommands[int(ic)]
}

type Request struct {
	Version  string
	Command  IPFSCommand
	Args     map[string]string
	FormArgs map[string]interface{}
}

type Response struct {
	Status int
	Msg    []byte
}

func createForm(buf *bytes.Buffer, args map[string]interface{}) (formType string, err error) {
	writer := multipart.NewWriter(buf)
	for k, v := range args {
		switch v.(type) {
		case *os.File:
			f, _ := v.(*os.File)
			fi, _ := f.Stat()
			formFile, err := writer.CreateFormFile(k, fi.Name())
			if err != nil {
				return "", err
			}
			_, err = io.Copy(formFile, f)
		default:
			err = writer.WriteField(k, v.(string))
			if err != nil {
				return
			}
		}
	}
	writer.Close()
	return writer.FormDataContentType(), nil
}

func (r *Request) Do() (result *Response, err error) {
	u := fmt.Sprintf("%s/%s/%s", IPFS_API_URL, r.Version, r.Command.String())
	ugs := url.Values{}
	for k, v := range r.Args {
		ugs.Add(k, v)
	}
	url, _ := url.Parse(u)
	url.RawQuery = ugs.Encode()

	buf := new(bytes.Buffer)
	formType, err := createForm(buf, r.FormArgs)
	if err != nil {
		return
	}
	fmt.Println(url.String())
	rsp, err := http.Post(url.String(), formType, buf)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	defer rsp.Body.Close()
	return &Response{Status: rsp.StatusCode, Msg: b}, nil
}

type Object struct {
	Name string
	Hash string
	Size string
}

func Add(filrName string) (obj *Object, err error) {
	r := Request{
		Version:  IPFS_VERSION_V0,
		Command:  ADD,
		Args:     make(map[string]string),
		FormArgs: make(map[string]interface{}),
	}
	r.Args["wrap-with-directory"] = "true"
	f, err := os.Open(filrName)
	if err != nil {
		return
	}
	defer f.Close()
	r.FormArgs["arg"] = f

	result, err := r.Do()
	if err != nil {
		return
	}

	if result.Status != 200 {
		fmt.Println(result)
		return nil, errors.New("request failed status= " + string(result.Status))
	}
	if len(result.Msg) == 0 {
		return
	}

	obj = &Object{}
	fmt.Println(string(result.Msg))
	err = json.Unmarshal(result.Msg, obj)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func Get(hash string) (data []byte, err error) {
	r := Request{
		Version:  IPFS_VERSION_V0,
		Command:  GET,
		Args:     make(map[string]string),
		FormArgs: make(map[string]interface{}),
	}
	r.Args["arg"] = hash
	r.Args["output"] = "./main.go"
	result, err := r.Do()
	if err != nil {
		return
	}
	if result.Status != 200 {
		fmt.Println(result)
		return nil, errors.New("request failed status= " + string(result.Status))
	}
	if len(result.Msg) == 0 {
		return
	}
	data = result.Msg
	MediaType, MediaParams, v := mime.ParseMediaType(string(data))
	fmt.Println(MediaType, MediaParams,v)
	return
}
