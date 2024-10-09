Test.rest("simple http", function(t)
	t.send("GET", "/")

	t.check(200, {
		message = "hello world"
	})
end)
