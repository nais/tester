Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"
)

Test.gql("test users", function(t)
	t.query [[
		query {
			users {
				id
				name
				email
			}
		}
	]]

	t.check {
		data = {
			users = {
				{
					id = Save("user_id"),
					name = "John Ddoe",
					email = Ignore(),
				}
			}
		}
	}
end)

Test.gql("get single user", function(t)
	t.query(string.format([[
		{
			user(id: "%s") {
				name
			}
		}
	]], State.user_id))

	t.check {
		data = {
			user = {
				name = "John Doe"
			}
		}
	}
end)
