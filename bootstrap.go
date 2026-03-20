package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const configFile = "config.json"

func ensureProjectFiles() error {
	configErr := ensureConfigFile()

	runtimeFiles := []struct {
		path    string
		content any
	}{
		{path: "profiles.json", content: map[string]UserProfile{}},
		{path: "signups.json", content: map[string]map[string][]string{}},
		{path: "test_signups.json", content: map[string]map[string][]string{}},
		{path: "admin_state.json", content: AdminState{
			AdminUsers:         map[string]bool{},
			BlockedSignupUsers: map[string]bool{},
		}},
		{path: "signup_schedule_state.json", content: SignupState{}},
	}

	for _, file := range runtimeFiles {
		if err := ensureJSONFile(file.path, file.content); err != nil {
			return err
		}
	}

	return configErr
}

func ensureConfigFile() error {
	if _, err := os.Stat(configFile); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("檢查 %s 失敗: %w", configFile, err)
	}

	if err := writeJSONFile(configFile, Config{}); err != nil {
		return fmt.Errorf("建立 %s 失敗: %w", configFile, err)
	}

	fmt.Printf("已建立 %s，接下來會在啟動流程中引導你完成初始化。\n", configFile)
	return nil
}

func ensureJSONFile(path string, value any) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("檢查 %s 失敗: %w", path, err)
	}

	if err := writeJSONFile(path, value); err != nil {
		return fmt.Errorf("建立 %s 失敗: %w", path, err)
	}

	return nil
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}
