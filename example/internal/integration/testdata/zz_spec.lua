-- This file is generated. Do not edit.


--- Ignore the field regardless of its value
---@return userdata
function Ignore()
  print("Ignore")
  ---@diagnostic disable-next-line: return-type-mismatch
  return {}
end

--- Ensure the field is not null, but allow any other value
---@return userdata
function NotNull()
	print("NotNull")
  ---@diagnostic disable-next-line: return-type-mismatch
	return {}
end

--- Save the field to the state. By default it will error if the field is null
---@param name string Name of the field in the state
---@param allowNull? boolean
---@return userdata
function Save(name, allowNull)
  print("Save: ", name, allowNull)
  ---@diagnostic disable-next-line: return-type-mismatch
  return {}
end

--- State variables
---@type table<string, any>
State = {}

--- Null ensures the value is null
---@type userdata
---@diagnostic disable-next-line: assign-type-mismatch
Null = {}

---@class TestFunctionTgql
local TestFunctionTgql = {}

--- Query comment
---@param query string
---@param headers? table
function TestFunctionTgql.query(query, headers)
  print("query")
end

--- Check comment
---@param resp table
function TestFunctionTgql.check(resp)
  print("check")
end

---@class TestFunctionTsql
local TestFunctionTsql = {}

--- Query for multiple rows
---@param query string
---@param ... string|boolean|number
function TestFunctionTsql.query(query, ...)
  print("query")
end

--- Query for a single row. Will error if no rows returned
---@param query string
---@param ... string|boolean|number
function TestFunctionTsql.queryRow(query, ...)
  print("queryRow")
end

--- Check comment
---@param resp table
function TestFunctionTsql.check(resp)
  print("check")
end

---@class TestFunctionTrest
local TestFunctionTrest = {}

--- Send http request
---@param method "GET" | "POST" | "PUT" | "DELETE" | "PATCH" | "OPTIONS" | "HEAD"
---@param path string
---@param body? string|table
function TestFunctionTrest.send(method, path, body)
  print("send")
end

--- Check the response done by send
---@param status_code number
---@param resp table
function TestFunctionTrest.check(status_code, resp)
  print("check")
end

--- Test case
---@class Test
---@field gql fun(name: string, fn: fun(t: TestFunctionTgql))
---@field sql fun(name: string, fn: fun(t: TestFunctionTsql))
---@field rest fun(name: string, fn: fun(t: TestFunctionTrest))
Test = {}

--- Helper functions
---@class Helper
Helper = {}

--- Execute some SQL. Will error if the SQL fails
---@param query string
---@param ... string|boolean|number
function Helper.SQLExec(query, ...)
  print("SQLExec")
end

--- Execute some SQL. Will return multiple rows.
---@param query string
---@param ... string|boolean|number
---@return table
function Helper.SQLQuery(query, ...)
  print("SQLQuery")
  return {}
end

--- Execute some SQL. Returns a single row. Error if no rows returned
---@param query string
---@param ... string|boolean|number
---@return table
function Helper.SQLQueryRow(query, ...)
  print("SQLQueryRow")
  return {}
end
