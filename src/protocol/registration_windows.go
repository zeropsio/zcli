//go:build windows

package protocol

import (
	"context"
	"fmt"
	"golang.org/x/sys/windows/registry"
)

type registration struct{}

func (r *registration) Register(ctx context.Context) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	rootKey, err := registry.OpenKey(registry.CLASSES_ROOT, "", registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open CLASSES_ROOT: %w", err)
	}
	defer rootKey.Close()

	protocolKey, _, err := registry.CreateKey(rootKey, "zcli", registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to create zcli key: %w", err)
	}
	defer protocolKey.Close()

	if err := protocolKey.SetStringValue("", "URL:zcli Protocol"); err != nil {
		return fmt.Errorf("failed to set default value: %w", err)
	}

	if err := protocolKey.SetStringValue("URL Protocol", ""); err != nil {
		return fmt.Errorf("failed to set URL Protocol value: %w", err)
	}

	defaultIconKey, _, err := registry.CreateKey(protocolKey, "DefaultIcon", registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to create DefaultIcon key: %w", err)
	}
	defer defaultIconKey.Close()

	iconPath := fmt.Sprintf("%s,1", execPath)
	if err := defaultIconKey.SetStringValue("", iconPath); err != nil {
		return fmt.Errorf("failed to set icon path: %w", err)
	}

	shellKey, _, err := registry.CreateKey(protocolKey, "shell", registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to create shell key: %w", err)
	}
	defer shellKey.Close()

	openKey, _, err := registry.CreateKey(shellKey, "open", registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to create open key: %w", err)
	}
	defer openKey.Close()

	commandKey, _, err := registry.CreateKey(openKey, "command", registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to create command key: %w", err)
	}
	defer commandKey.Close()

	commandValue := fmt.Sprintf("\"%s\" protocol handle \"%%1\"", execPath)
	if err := commandKey.SetStringValue("", commandValue); err != nil {
		return fmt.Errorf("failed to set command value: %w", err)
	}

	return nil
}

func (r *registration) Unregister(ctx context.Context) error {
	rootKey, err := registry.OpenKey(registry.CLASSES_ROOT, "", registry.CREATE_SUB_KEY)
	if err != nil {
		return fmt.Errorf("failed to open CLASSES_ROOT: %w", err)
	}
	defer rootKey.Close()

	if err := registry.DeleteKey(rootKey, "zcli"); err != nil {
		if err != registry.ErrNotExist {
			return fmt.Errorf("failed to delete zcli key: %w", err)
		}
	}

	return nil
}

func (r *registration) IsRegistered(ctx context.Context) (bool, error) {
	rootKey, err := registry.OpenKey(registry.CLASSES_ROOT, "", registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("failed to open CLASSES_ROOT: %w", err)
	}
	defer rootKey.Close()

	protocolKey, err := registry.OpenKey(rootKey, "zcli", registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to open zcli key: %w", err)
	}
	defer protocolKey.Close()

	_, _, err = protocolKey.GetStringValue("URL Protocol")
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to read URL Protocol value: %w", err)
	}

	return true, nil
}