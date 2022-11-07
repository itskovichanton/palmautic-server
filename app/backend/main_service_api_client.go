package backend

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/goava/pkg/goava/utils/case_insensitive"
	"github.com/spf13/cast"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const SqlParamTypeSuffix = "___JAVASVCTYPE"
const SqlParamIsNull = "___ISNULL"

type IMainServiceAPIClientService interface {
	QueryDomainDBForMaps(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	QueryDomainDBForMap(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	Query(mainServiceUrl string, uri string, sql string, params map[string]interface{}, resultGetter func() interface{}, dataGetter func(interface{}) (interface{}, error), paramTypes map[string]string) (*Query, error)
	QueryDomainDBForMapsWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	QueryDomainDBForMapWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	ExecDomainDBWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	ExecDomainDB(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	UpdateDomainDBWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
	UpdateDomainDB(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error)
}

type Query struct {
	Details, Sql string
	Result       interface{}
}

type JavaMainServiceAPIClientServiceImpl struct {
	IMainServiceAPIClientService

	Config     *core.Config
	HttpClient *http.Client
	Generator  goava.IGenerator
}

type BException struct {
	Name          string
	DetailMessage string
}

type IntercomExceptionData struct {
	Stack string
	Code  string
	//Chain []string
}

type IntercomServiceError struct {
	errs.BaseError

	Data IntercomExceptionData
}

func (c *JavaMainServiceAPIClientServiceImpl) QueryDomainDBForMaps(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.QueryDomainDBForMapsWithSpecifiedUrl(c.Config.Props.MainServiceUrl, sql, params, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) QueryDomainDBForMap(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.QueryDomainDBForMapWithSpecifiedUrl(c.Config.Props.MainServiceUrl, sql, params, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) ExecDomainDB(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.ExecDomainDBWithSpecifiedUrl(c.Config.Props.MainServiceUrl, sql, params, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) UpdateDomainDB(sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.UpdateDomainDBWithSpecifiedUrl(c.Config.Props.MainServiceUrl, sql, params, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) QueryDomainDBForMapsWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.Query(mainServiceUrl, "domaindb/query", sql, params, func() interface{} {
		return []map[string]interface{}{}
	}, func(data interface{}) (interface{}, error) {
		return utils.ToSliceStringMapE(data)
	}, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) QueryDomainDBForMapWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.Query(mainServiceUrl, "domaindb/queryfirst", sql, params, func() interface{} {
		return map[string]interface{}{}
	}, func(data interface{}) (interface{}, error) {
		return cast.ToStringMapE(data)
	}, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) ExecDomainDBWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.Query(mainServiceUrl, "domaindb/executeProc", sql, params, func() interface{} {
		return map[string]interface{}{}
	}, func(data interface{}) (interface{}, error) {
		return cast.ToStringMapE(data)
	}, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) UpdateDomainDBWithSpecifiedUrl(mainServiceUrl string, sql string, params map[string]interface{}, paramTypes map[string]string) (*Query, error) {
	return c.Query(mainServiceUrl, "domaindb/update", sql, params, func() interface{} {
		return map[string]interface{}{}
	}, func(data interface{}) (interface{}, error) {
		return cast.ToStringMapE(data)
	}, paramTypes)
}

func (c *JavaMainServiceAPIClientServiceImpl) Query(mainServiceUrl string, uri string, sql string, params map[string]interface{}, resultGetter func() interface{}, dataGetter func(interface{}) (interface{}, error), paramTypes map[string]string) (*Query, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	r, err := c.query(mainServiceUrl, uri, sql, params, resultGetter, dataGetter, paramTypes)
	if err != nil {
		return r, &MainServiceApiClientError{
			BaseError: *errs.NewBaseErrorFromCause(err),
			Params:    params,
		}
	}

	return r, nil
}

func (c *JavaMainServiceAPIClientServiceImpl) query(mainServiceUrl string, uri string, sql string, params map[string]interface{}, resultGetter func() interface{}, dataGetter func(interface{}) (interface{}, error), paramTypes map[string]string) (*Query, error) {

	if paramTypes == nil {
		paramTypes = map[string]string{}
	}
	params["sql"] = sql

	request, err := c.createRequest(mainServiceUrl+"/"+uri, params, paramTypes)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		intercomMethod := resp.Header.Get("INTERCOM-METHOD")
		if strings.EqualFold("json", intercomMethod) {
			var data IntercomExceptionData
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				return nil, err
			}
			return nil, &IntercomServiceError{
				BaseError: *errs.NewBaseErrorWithReasonDetails("", "INTERCOM ERROR. Code="+data.Code, errs.ReasonOther),
				Data:      data,
			}
		}

		return nil, errors.New(fmt.Sprintf("Intercom-Method %v is not supported", intercomMethod))
	}

	result := Query{}
	result.Result = resultGetter()

	dec := json.NewDecoder(resp.Body)
	//dec.UseNumber()
	err = dec.Decode(&result)
	if err != nil {
		if err.Error() == "EOF" {
			return &result, nil
		}
		return nil, err
	}

	if dataGetter != nil && result.Result != nil {
		result.Result, err = dataGetter(result.Result)
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
}

func (c *JavaMainServiceAPIClientServiceImpl) toJavaType(v interface{}, paramName string, paramTypes map[string]interface{}) string {

	if paramTypes != nil {
		paramType := case_insensitive.Get(paramTypes, paramName)
		if paramType != nil {
			if strings.EqualFold("date", cast.ToString(paramType)) {
				return "java.util.Date"
			}
		}
	}

	switch v.(type) {
	case *os.File:
		return "byte[]"
	case []byte:
		return "byte[]"
	case int:
		return "java.lang.Integer"
	case string:
		return "java.lang.String"
	case int64:
		return "java.lang.Long"
	case float32:
		return "java.lang.Double"
	default:
		return "java.lang.Object"
	}
}

func (c *JavaMainServiceAPIClientServiceImpl) createRequest(url string, params map[string]interface{}, paramTypes map[string]string) (*http.Request, error) {

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var valueReader io.Reader

	for k, v := range paramTypes {
		params[k+SqlParamTypeSuffix] = c.sqlToJavaType(v)
	}

	for k, v := range params {

		var fw io.Writer
		var err error

		switch vt := v.(type) {
		case *multipart.FileHeader:

			file, err := vt.Open()
			if err != nil {
				return nil, err
			}

			savedMultipartFormFilePath, err := c.Config.GetTempFile("multipart_" + vt.Filename + "_*")

			buf, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(savedMultipartFormFilePath.Name(), buf, os.ModePerm)
			if err != nil {
				return nil, err
			}

			valueReader = savedMultipartFormFilePath
			v = savedMultipartFormFilePath
		}

		if strings.Index(k, SqlParamTypeSuffix) == -1 {
			if fw, err = w.CreateFormField(k + SqlParamTypeSuffix); err != nil {
				return nil, err
			}
			if _, err = io.Copy(fw, strings.NewReader(c.toJavaType(v, k, utils.ToStringMap(paramTypes)))); err != nil {
				return nil, err
			}

			if v == nil {
				if fw, err = w.CreateFormField(k + SqlParamIsNull); err != nil {
					return nil, err
				}
				if _, err = io.Copy(fw, strings.NewReader("true")); err != nil {
					return nil, err
				}
			}
		}

		switch value := v.(type) {
		case []byte:

			ext := paramTypes[k]
			fileName := fmt.Sprintf("_msac_field_%v-%v", k, c.Generator.GenerateUint64())
			if len(ext) > 0 {
				fileName += "." + ext
			}

			v, err = utils.WriteToFile(filepath.Join(c.Config.GetTempFilesStorageDir(), fileName), func(w *bufio.Writer) error {
				_, err := w.Write(value)
				return err
			})
			if err != nil {
				return nil, err
			}
		}

		switch value := v.(type) {
		case *os.File:
			value, err = os.Open(value.Name())
			valueReader = bufio.NewReader(value)
			defer utils.CloseAndRemove(value)
		default:
			valueReader = strings.NewReader(fmt.Sprintf("%v", v))
		}

		if x, ok := v.(io.Closer); ok {
			defer x.Close()
		}

		if x, ok := v.(*os.File); ok {
			if fw, err = w.CreateFormFile(k, x.Name()); err != nil {
				return nil, err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(k); err != nil {
				return nil, err
			}
		}
		if _, err = io.Copy(fw, valueReader); err != nil {
			return nil, err
		}

	}

	w.Close()

	r, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return r, nil
	}
	r.Header.Set("Content-Type", w.FormDataContentType())
	//request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("INTERCOM-METHOD", "json")
	r.Header.Set("intercom-params-ignore-err", "true")

	return r, nil
}

func (c *JavaMainServiceAPIClientServiceImpl) sqlToJavaType(v string) interface{} {
	if strings.EqualFold(v, "varchar") {
		return "java.lang.String"
	}
	if strings.EqualFold("date", v) {
		return "java.util.Date"
	}
	return "java.lang.Object"
}

type MainServiceApiClientError struct {
	errs.BaseError

	Params map[string]interface{}
}

func (c *MainServiceApiClientError) Error() string {
	return c.BaseError.Error() + " [QUERY: " + utils.ToJson(c.Params) + "]"
}
