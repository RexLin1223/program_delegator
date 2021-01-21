package windows

import (
	"errors"
	"golang.org/x/sys/windows/registry"
)

var RegCategoryMap = map[string]registry.Key {
	"HKEY_CLASSES_ROOT": registry.CLASSES_ROOT,
	"HKEY_CURRENT_USER": registry.CURRENT_USER,
	"HKEY_LOCAL_MACHINE": registry.LOCAL_MACHINE,
	"HKEY_USERS": registry.USERS,
	"HKEY_CURRENT_CONFIG": registry.CURRENT_CONFIG,
	"HKEY_PERFORMANCE_DATA": registry.PERFORMANCE_DATA,
}

func QueryRegKey64(keyCategory string, keyPath string, key string ) (string, error){
	cat, ok := RegCategoryMap[keyCategory]
	if !ok  {
		return "", errors.New("invalid key category")
	}

	r, err:=registry.OpenKey(cat, keyPath, registry.ALL_ACCESS | registry.WOW64_64KEY)
	if err!=nil {
		return "", err
	}

	defer r.Close()
	s, _, err:=r.GetStringValue(key)
	if err!= nil{
		return "", err
	}
	return s, nil
}

func QueryRegKey32(keyCategory string, keyPath string, key string )(string, error){
	cat, ok := RegCategoryMap[keyCategory]
	if !ok  {
		return "", errors.New("invalid key category")
	}

	r, err:=registry.OpenKey(cat, keyPath, registry.ALL_ACCESS | registry.WOW64_32KEY)
	if err!=nil {
		return "", err
	}

	defer r.Close()
	s, _, err:=r.GetStringValue(key)
	if err!= nil{
		return "", err
	}
	return s, nil
}