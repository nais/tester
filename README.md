# Integration test framework

This package contains a custom integration framework that let's you test your Go application using a set of Lua tests.
There's support to test the system using:

- REST
- GraphQL
- SQL
- PubSub

Other types of tests can be added by implementing the `Runner` interface.

## Setup

For tester to work, a manager has to be created.
This is where the runners and capabilities are registered.

See [example/internal/integration/manager.go](./example/internal/integration/manager.go) for full example.

To run the tests, create a `_test.go` file containing the following:

```go
package integration
	
import (
	"context"
	"testing"

	"github.com/nais/tester/lua"
)

func TestIntegration(t *testing.T) {
	mgr, err := TestRunner(false)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := mgr.Run(ctx, "./testdata", lua.NewTestReporter(t)); err != nil {
		t.Fatal(err)
	}
}
```

This will run all tests within the `testdata` folder.

### Generate Lua spec

To help with writing tests, you can generate a Lua spec file from the Go code.
This will generate a Lua file with all the available functions and their parameters.

This also requires a manager to be created.
See [example/internal/tools/tester_spec/main.go](./example/internal/tools/tester_spec/main.go) for full example.

### Graphical UI

When writing tests, rerunning the entire test suite can be time consuming.
To help with the case where there's only Lua tests that are changed, a graphical UI can be used to run the tests.

See [example/internal/tools/tester_run/main.go](./example/internal/tools/tester_run/main.go) for full example.
The `--ui` flag will start a web server that can be accessed at `http://localhost:9876`.

## How to use it

Within the test folder, every Lua file is considered a standalone environment.
The tests are run in order of declaration.

### State

You can store state between tests using the `State` object.
By saving state, use `Save(name)` when comparing the expected result with the actual result.

```lua
Test.gql("test users", function(t)
  t.query "..."
  t.check {
    data = {
      users = {
        { id = Save("user1") },
        { id = Save("user2") },
      }
    }
  }
end)

Test.gql("test user", function(t)
  t.query(string.format([[{ user(id: "%s") { ... } }]], State.user1))
  t.check { /* ... */ }
end)
```

### Nil checks

When using `nil` in lua, the field is removed when comparing the results.
If you want to check that a field is `nil`, use the `Null` object.

### Ignoring fields

When comparing the expected results and the field is dynamic, you can use the `Ignore` function.
This will ignore the field when comparing the results.
If you want to check if the field is set, use the `NotNull()` function.

```lua
Test.rest("test users", function(t)
  t.send("GET", "/users")
  t.check {
    data = {
      users = {
        { id = Ignore(), name = "John" },
        { id = NotNull(), name = "Jane" },
      }
    }
  }
end)
```

### String contains

When comparing strings, you can use the `Contains` function.
This will check if the string contains the given substring.

```lua
Test.rest("test users", function(t)
  t.send("GET", "/users")
  t.check {
    data = {
      users = {
        { id = Ignore(), name = Contains("John") },
        { id = NotNull(), name = "Jane" },
      }
    }
  }
end)
```

## Configuration

A configuration struct can be used to allow each test to have different configurations.
The configuration is set on the `Config` object, and should be set as the first thing in the test.

```lua
Config.seedDB = true
Config.loadData = "./data.json"

Test.gql("my test", function(t)
  -- ...
```

## Runners

### Graphql

The GraphQL runner can be used like this:

```lua
Test.gql("test users", function(t)
  t.addHeader("Authorization", "Bearer token")

  t.query [[{ users { id name } }]]

  t.check {
    data = {
      users = {
        { id = Ignore(), name = "John" },
        { id = NotNull(), name = "Jane" },
      }
    }
  }
end)
```

### REST

The REST runner can be used like this:

```lua
Test.rest("test users", function(t)
  t.send("POST", "/users", { name = "John" })
  t.check {
    data = {
      users = {
        { id = Ignore(), name = "John" },
        { id = NotNull(), name = "Jane" },
      }
    }
  }
end)
```

### SQL

The SQL runner can be used like this:

```lua
Test.sql("test users", function(t)
  t.query("SELECT * FROM users")
  t.check {
    { id = Ignore(), name = "John" },
    { id = NotNull(), name = "Jane" },
  }
end)

Test.sql("test single users", function(t)
  t.query("SELECT * FROM users WHERE id = 1")
  t.check {
    id = Ignore(),
    name = "John",
  }
end)
```

#### Helpers

`Helper.SQLExec(q, ...)` can be used to execute a SQL query.

`Helper.SQLQueryRow(q, ...)` can be used to query a single row.

`Helper.SQLQuery(q, ...)` can be used to query multiple rows.

### PubSub

The PubSub runner checks if a message matches what is expected.
The runner will not wait for the message to be received, so the test should be run after the message is sent.

```lua
Test.pubsub("test users", function(t)
  t.check "my-topic", {
    id = Ignore(),
    name = "John",
  }
end)
```

#### Helpers

`Helper.emptyPubSubTopic(topic)` can be used to empty a topic.

## Code generated by GitHub Copilot

This repository uses GitHub Copilot to generate code.
