package validation

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/robfig/cron/v3"
)

// IsValidHostnameOrIP проверяет, допустим ли hostname или IP-адрес
func IsValidHostnameOrIP(host string, fieldName string, isAllowEmpty bool) (bool, error) {
	if host == "" {
		if isAllowEmpty {
			return true, nil
		}
		return false, fmt.Errorf("%s не может быть пустым", fieldName)
	}

	// Use net.ParseIP to try to parse the address.
	if ip := net.ParseIP(host); ip != nil {
		return true, nil
	}

	// Manual check for IPv4 addresses (since ParseIP can miss cases like "256.256.256.256").
	if ip := net.ParseIP(host); ip == nil {
		parts := strings.Split(host, ".")
		if len(parts) == 4 {
			for _, part := range parts {
				if num, err := strconv.Atoi(part); err == nil {
					if num < 0 || num > 255 {
						return false, fmt.Errorf("%s '%s' недействителен", fieldName, host)
					}
				} else {
					return false, fmt.Errorf("%s '%s' недействителен", fieldName, host)
				}
			}
			return true, nil
		}
	}

	// Validate the hostname
	if len(host) > 255 {
		return false, fmt.Errorf("%s '%s' недействителен", fieldName, host)
	}
	for _, part := range strings.Split(host, ".") {
		if len(part) == 0 || len(part) > 63 {
			return false, fmt.Errorf("%s '%s' недействителен", fieldName, host)
		}
		if part[0] == '-' || part[len(part)-1] == '-' {
			return false, fmt.Errorf("%s '%s' недействителен", fieldName, host)
		}
		for _, char := range part {
			if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != '-' {
				return false, fmt.Errorf("%s '%s' недействителен", fieldName, host)
			}
		}
	}

	return true, nil
}

func IsValidPort(port int, fieldName string, isAllowZero bool) (bool, error) {
	if isAllowZero && port == 0 {
		return true, nil
	}
	if port <= 0 || port > 65535 {
		return false, fmt.Errorf("%s '%d' недействителен, должен быть в диапазоне %d-%d", fieldName, port, 1, 65535)
	}
	return true, nil
}

func IsValidValueInList(value string, fieldName string, allowed []string, isCaseSensitive bool) (bool, error) {
	if !isCaseSensitive {
		value = strings.ToLower(value)
		for i, val := range allowed {
			allowed[i] = strings.ToLower(val)
		}
	}

	for _, val := range allowed {
		if value == val {
			return true, nil
		}
	}
	return false, fmt.Errorf("%s '%s' недействителен, должен быть один из: %v", fieldName, value, allowed)
}

func IsNotEmpty(value string, fieldName string) (bool, error) {
	if value == "" {
		return false, fmt.Errorf("%s не может быть пустым", fieldName)
	}
	return true, nil
}

func IsValidStringLength(value string, fieldName string, minLen, maxLen int, isAllowEmpty bool) (bool, error) {
	length := len(value)
	if length == 0 {
		if isAllowEmpty {
			return true, nil
		}
		return false, fmt.Errorf("%s не может быть пустым", fieldName)
	}
	if length < minLen || (maxLen > 0 && length > maxLen) {
		return false, fmt.Errorf("%s '%s' недействительна, длина должна быть в диапазоне %d-%d символов", fieldName, value, minLen, maxLen)
	}
	return true, nil
}

func IsAlphanumeric(value string, fieldName string, isAllowEmpty bool) (bool, error) {
	if len(value) == 0 {
		if isAllowEmpty {
			return true, nil
		}
		return false, fmt.Errorf("%s не может быть пустым", fieldName)
	}
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return false, fmt.Errorf("%s '%s' недействителен, должен содержать только буквы и цифры", fieldName, value)
		}
	}
	return true, nil
}

func IsValidURL(value string, fieldName string, isAllowEmpty bool) (bool, error) {
	if len(value) == 0 {
		if isAllowEmpty {
			return true, nil
		}
		return false, fmt.Errorf("%s не может быть пустым", fieldName)
	}
	_, err := url.ParseRequestURI(value)
	if err != nil {
		return false, fmt.Errorf("%s '%s' недействителен как URL: %v", fieldName, value, err)
	}
	return true, nil
}

func IsValidEmail(value string, fieldName string, isAllowEmpty bool) (bool, error) {
	if len(value) == 0 {
		if isAllowEmpty {
			return true, nil
		}
		return false, fmt.Errorf("%s не может быть пустым", fieldName)
	}
	var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(value) {
		return false, fmt.Errorf("%s '%s' недействителен как email", fieldName, value)
	}
	return true, nil
}

func IsPositiveInt(value int, fieldName string) (bool, error) {
	if value <= 0 {
		return false, fmt.Errorf("%s '%d' недействителен, должен быть положительным числом", fieldName, value)
	}
	return true, nil
}

// IsValidCronExpression проверяет, является ли cron выражение валидным.
func IsValidCronExpression(expression string, fieldName string, isWithSeconds bool) (bool, error) {
	var parser cron.Parser
	if isWithSeconds {
		parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	} else {
		parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	}

	_, err := parser.Parse(expression)
	if err != nil {
		return false, fmt.Errorf("%s cron выражение '%s' недействительно: %v", fieldName, expression, err)
	}
	return true, nil
}

// IsValidRobfigCronDescriptor проверяет валидность спец-выражений cron библиотеки robfig/cron (например, "@every 2m", "@hourly").
func IsValidRobfigCronDescriptor(expression string, fieldName string) (bool, error) {
	_, err := cron.ParseStandard(expression)
	if err != nil {
		return false, fmt.Errorf("%s cron выражение '%s' недействительно: %v", fieldName, expression, err)
	}
	return true, nil
}
