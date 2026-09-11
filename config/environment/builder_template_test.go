package environment

import (
	"testing"
)

func TestEnvironmentTemplate(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.url":  "postgres://localhost:5432/test",
		"database.host": "localhost",
		"database.port": 5432,
	}))

	template := NewEnvironmentTemplate(env)

	url := template.GetDatabaseURL("default")
	if url != "postgres://localhost:5432/test" {
		t.Errorf("expected database URL 'postgres://localhost:5432/test', got %s", url)
	}

	host := template.GetDatabaseHost("default")
	if host != "localhost" {
		t.Errorf("expected database host 'localhost', got %s", host)
	}

	port := template.GetDatabasePort(0)
	if port != 5432 {
		t.Errorf("expected database port 5432, got %d", port)
	}
}

func TestEnvironmentTemplate_GetDatabaseName(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.name": "mydb",
	}))

	template := NewEnvironmentTemplate(env)
	name := template.GetDatabaseName("default")

	if name != "mydb" {
		t.Errorf("expected 'mydb', got %s", name)
	}
}

func TestEnvironmentTemplate_GetServerHost(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.host": "0.0.0.0",
	}))

	template := NewEnvironmentTemplate(env)
	host := template.GetServerHost("default")

	if host != "0.0.0.0" {
		t.Errorf("expected '0.0.0.0', got %s", host)
	}
}

func TestEnvironmentTemplate_GetServerPort(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port": 8080,
	}))

	template := NewEnvironmentTemplate(env)
	port := template.GetServerPort(3000)

	if port != 8080 {
		t.Errorf("expected 8080, got %d", port)
	}
}

func TestEnvironmentTemplate_GetLogLevel(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"log.level": "debug",
	}))

	template := NewEnvironmentTemplate(env)
	level := template.GetLogLevel("info")

	if level != "debug" {
		t.Errorf("expected 'debug', got %s", level)
	}
}

func TestEnvironmentTemplate_GetRedisHost(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.host": "127.0.0.1",
	}))

	template := NewEnvironmentTemplate(env)
	host := template.GetRedisHost("localhost")

	if host != "127.0.0.1" {
		t.Errorf("expected '127.0.0.1', got %s", host)
	}
}

func TestEnvironmentTemplate_GetRedisPort(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.port": 6379,
	}))

	template := NewEnvironmentTemplate(env)
	port := template.GetRedisPort(6380)

	if port != 6379 {
		t.Errorf("expected 6379, got %d", port)
	}
}

func TestEnvironmentTemplate_GetRedisPassword(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.password": "secret",
	}))

	template := NewEnvironmentTemplate(env)
	password := template.GetRedisPassword("")

	if password != "secret" {
		t.Errorf("expected 'secret', got %s", password)
	}
}

func TestEnvironmentTemplate_IsDebugMode(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"debug": true,
	}))

	template := NewEnvironmentTemplate(env)
	if !template.IsDebugMode() {
		t.Error("expected debug mode to be true")
	}
}

func TestEnvironmentTemplate_IsVerbose(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"verbose": true,
	}))

	template := NewEnvironmentTemplate(env)
	if !template.IsVerbose() {
		t.Error("expected verbose mode to be true")
	}
}
