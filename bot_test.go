package telego

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mymmrac/go-telegram-bot-api/api"
)

const (
	token        = "1234567890:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc"
	invalidToken = "invalid-token"

	methodName = "testMethod"
)

var errTest = errors.New("error")

func Test_validateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		isValid bool
	}{
		{
			name:    "empty",
			token:   "",
			isValid: false,
		},
		{
			name:    "not valid",
			token:   invalidToken,
			isValid: false,
		},
		{
			name:    "valid 1",
			token:   token,
			isValid: true,
		},
		{
			name:    "valid 2",
			token:   "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc",
			isValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateToken(tt.token)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestNewBot(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bot, err := NewBot(token)

		assert.NoError(t, err)
		assert.NotNil(t, bot)
	})

	t.Run("success with options", func(t *testing.T) {
		bot, err := NewBot(token, func(_ *Bot) error { return nil })

		assert.NoError(t, err)
		assert.NotNil(t, bot)
	})

	t.Run("error", func(t *testing.T) {
		bot, err := NewBot(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, bot)
	})

	t.Run("error with options", func(t *testing.T) {
		bot, err := NewBot(token, func(_ *Bot) error { return errTest })

		assert.ErrorIs(t, err, errTest)
		assert.Nil(t, bot)
	})
}

func TestBot_Token(t *testing.T) {
	bot, err := NewBot(token)
	assert.NoError(t, err)

	assert.Equal(t, token, bot.Token())
}

func Test_parseParameters(t *testing.T) {
	n := 1

	tests := []struct {
		name             string
		parameters       interface{}
		parsedParameters map[string]string
		isError          bool
	}{
		{
			name: "success",
			parameters: &struct {
				Empty  string `json:"empty,omitempty"`
				Number int    `json:"number"`
				Array  []int  `json:"array"`
				Struct *struct {
					N int `json:"n"`
				} `json:"struct"`
			}{
				Number: 10,
				Array:  []int{1, 2, 3},
				Struct: &struct {
					N int `json:"n"`
				}{2},
			},
			parsedParameters: map[string]string{
				"number": "10",
				"array":  "[1,2,3]",
				"struct": "{\"n\":2}",
			},
			isError: false,
		},
		{
			name: "error not pointer",
			parameters: struct {
				a int
			}{},
			parsedParameters: nil,
			isError:          true,
		},
		{
			name:             "error not struct",
			parameters:       &n,
			parsedParameters: nil,
			isError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedParameters, err := parseParameters(tt.parameters)
			if tt.isError {
				assert.Error(t, err)
				assert.Nil(t, parsedParameters)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.parsedParameters, parsedParameters)
		})
	}
}

var testFile = &os.File{}

type testStruct struct{}

func (ts *testStruct) fileParameters() map[string]*os.File {
	return map[string]*os.File{
		"test": testFile,
	}
}

func Test_filesParameters(t *testing.T) {
	tests := []struct {
		name       string
		parameters interface{}
		files      map[string]*os.File
		hasFiles   bool
	}{
		{
			name:       "with files",
			parameters: &testStruct{},
			files: map[string]*os.File{
				"test": testFile,
			},
			hasFiles: true,
		},
		{
			name:       "no files",
			parameters: 1,
			files:      nil,
			hasFiles:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, hasFiles := filesParameters(tt.parameters)
			assert.Equal(t, tt.hasFiles, hasFiles)
			assert.Equal(t, tt.files, files)
		})
	}
}

type paramsWithFile struct {
	N int `json:"n"`
}

func (p *paramsWithFile) fileParameters() map[string]*os.File {
	return map[string]*os.File{
		"test": {},
	}
}

type notStructParamsWithFile string

func (p *notStructParamsWithFile) fileParameters() map[string]*os.File {
	return map[string]*os.File{
		"test": {},
	}
}

func TestBot_constructAndCallRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mb := newMockedBot(ctrl)

	params := struct {
		N int `json:"n"`
	}{
		N: 1,
	}

	url := mb.Bot.apiURL + "/bot" + mb.Bot.token + "/" + methodName

	expectedResp := &api.Response{
		Ok: true,
	}

	paramsBytes, err := json.Marshal(params)
	assert.NoError(t, err)

	expectedData := &api.RequestData{
		ContentType: api.ContentTypeJSON,
		Buffer:      bytes.NewBuffer(paramsBytes),
	}

	t.Run("success json", func(t *testing.T) {
		mb.MockRequestConstructor.EXPECT().
			JSONRequest(params).
			Return(expectedData, nil).
			Times(1)

		mb.MockAPICaller.EXPECT().
			Call(url, expectedData).
			Return(expectedResp, nil).
			Times(1)

		resp, err := mb.Bot.constructAndCallRequest(methodName, params)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("error json", func(t *testing.T) {
		mb.MockRequestConstructor.EXPECT().
			JSONRequest(params).
			Return(nil, errTest).
			Times(1)

		resp, err := mb.Bot.constructAndCallRequest(methodName, params)
		assert.ErrorIs(t, err, errTest)
		assert.Nil(t, resp)
	})

	t.Run("success multipart", func(t *testing.T) {
		paramsFile := &paramsWithFile{N: 1}
		paramsMap := map[string]string{
			"n": "1",
		}

		paramsBytesFile, err := json.Marshal(paramsFile)
		assert.NoError(t, err)

		expectedDataFile := &api.RequestData{
			ContentType: api.ContentTypeJSON,
			Buffer:      bytes.NewBuffer(paramsBytesFile),
		}

		mb.MockRequestConstructor.EXPECT().
			MultipartRequest(paramsMap, gomock.Any()).
			Return(expectedDataFile, nil).
			Times(1)

		mb.MockAPICaller.EXPECT().
			Call(url, expectedDataFile).
			Return(expectedResp, nil).
			Times(1)

		resp, err := mb.Bot.constructAndCallRequest(methodName, paramsFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("error multipart", func(t *testing.T) {
		paramsFile := &paramsWithFile{N: 1}
		paramsMap := map[string]string{
			"n": "1",
		}

		mb.MockRequestConstructor.EXPECT().
			MultipartRequest(paramsMap, gomock.Any()).
			Return(nil, errTest).
			Times(1)

		resp, err := mb.Bot.constructAndCallRequest(methodName, paramsFile)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("error multipart params", func(t *testing.T) {
		notStruct := notStructParamsWithFile("test")

		resp, err := mb.Bot.constructAndCallRequest(methodName, &notStruct)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("error call", func(t *testing.T) {
		mb.MockRequestConstructor.EXPECT().
			JSONRequest(params).
			Return(expectedData, nil).
			Times(1)

		mb.MockAPICaller.EXPECT().
			Call(url, expectedData).
			Return(nil, errTest).
			Times(1)

		resp, err := mb.Bot.constructAndCallRequest(methodName, params)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBot_performRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mb := newMockedBot(ctrl)

	params := struct {
		N int `json:"n"`
	}{
		N: 1,
	}

	var result int

	t.Run("success", func(t *testing.T) {
		mb.MockRequestConstructor.EXPECT().
			JSONRequest(gomock.Any()).
			Return(&api.RequestData{}, nil).
			Times(1)

		mb.MockAPICaller.EXPECT().
			Call(gomock.Any(), gomock.Any()).
			Return(&api.Response{
				Ok:     true,
				Result: bytes.NewBufferString("1").Bytes(),
				Error:  nil,
			}, nil)

		err := mb.Bot.performRequest(methodName, params, &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("error not ok", func(t *testing.T) {
		mb.MockRequestConstructor.EXPECT().
			JSONRequest(gomock.Any()).
			Return(&api.RequestData{}, nil).
			Times(1)

		mb.MockAPICaller.EXPECT().
			Call(gomock.Any(), gomock.Any()).
			Return(&api.Response{
				Ok:     false,
				Result: nil,
				Error:  &api.Error{},
			}, nil)

		err := mb.Bot.performRequest(methodName, params, &result)
		assert.Error(t, err)
	})

	t.Run("error construct and call", func(t *testing.T) {
		mb.MockRequestConstructor.EXPECT().
			JSONRequest(gomock.Any()).
			Return(nil, errTest).
			Times(1)

		err := mb.Bot.performRequest(methodName, params, &result)
		assert.Error(t, err)
	})
}
