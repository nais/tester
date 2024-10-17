Test.rest("simple http", function(t)
	t.send("GET", "/")

	t.check(200, {
		message = "hello world",
		time = Ignore() -- Ignore succeeds as the response does not contain a time field
		-- updated = NotNull() -- This would fail as the response does not contain an updated field
	})

	t.check(200, {
		message = Contains("hello"),
		time = Ignore()
	})

	t.check(200, {
		message = Contains("Hello", false),
		time = Ignore()
	})
end)
