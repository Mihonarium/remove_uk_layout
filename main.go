package main

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"log"
)

func main() {
	if err := deleteKeyboardLayout("00000809"); err != nil {
		log.Fatal(err)
	}
	if err := deleteKeyboardLayout("0x00000809"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Finished deleting keyboard layout entries. Please sign out & sign back in or restart your computer.")

	// wait for user input to exit
	fmt.Println("Press enter to exit.")
	_, _ = fmt.Scanln()
}

func deleteKeyboardLayout(keyboardID string) error {
	// Delete from Substitutes and get the substitute IDs
	substituteIDs, err := deleteFromRegistryAndGetSubstitutes(registry.CURRENT_USER, `Keyboard Layout\Substitutes`, keyboardID)
	if err != nil {
		return fmt.Errorf("error deleting from Substitutes: %w", err)
	}

	// Delete from Preload using both the keyboardID and substitute IDs
	if err := deleteFromPreload(registry.CURRENT_USER, `Keyboard Layout\Preload`, keyboardID, substituteIDs); err != nil {
		return fmt.Errorf("error deleting from Preload: %w", err)
	}

	return nil
}

func deleteFromRegistryAndGetSubstitutes(baseKey registry.Key, subKeyPath, keyboardID string) ([]string, error) {
	var substituteIDs []string

	key, err := registry.OpenKey(baseKey, subKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return nil, fmt.Errorf("unable to open registry key: %w", err)
	}
	defer func(key registry.Key) {
		err := key.Close()
		if err != nil {
			fmt.Println("error closing registry key:", err)
		}
	}(key)

	values, err := key.ReadValueNames(0)
	if err != nil {
		return nil, fmt.Errorf("error reading values: %w", err)
	}

	for _, value := range values {
		data, _, err := key.GetStringValue(value)
		if err != nil {
			return nil, fmt.Errorf("error reading value data: %w", err)
		}

		if data == keyboardID {
			substituteIDs = append(substituteIDs, value)
			if err := key.DeleteValue(value); err != nil {
				return nil, fmt.Errorf("error deleting value: %w", err)
			}
			fmt.Printf("Deleted %s\\%s\n", subKeyPath, value)
		}
	}

	return substituteIDs, nil
}

func deleteFromPreload(baseKey registry.Key, subKeyPath, keyboardID string, substituteIDs []string) error {
	key, err := registry.OpenKey(baseKey, subKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("unable to open registry key: %w", err)
	}
	defer key.Close()

	values, err := key.ReadValueNames(0)
	if err != nil {
		return fmt.Errorf("error reading values: %w", err)
	}

	for _, value := range values {
		data, _, err := key.GetStringValue(value)
		if err != nil {
			return fmt.Errorf("error reading value data: %w", err)
		}

		if data == keyboardID || contains(substituteIDs, data) ||
			contains(substituteIDs, fmt.Sprintf("0x%s", data)) || contains(substituteIDs, value) ||
			contains(substituteIDs, fmt.Sprintf("0x%s", value)) {
			if err := key.DeleteValue(value); err != nil {
				return fmt.Errorf("error deleting value: %w", err)
			}
			fmt.Printf("Deleted %s\\%s\n", subKeyPath, value)
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
