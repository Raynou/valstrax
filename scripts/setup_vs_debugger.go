package main

import "os"

func main() {
	content := `{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "go",
      "request": "launch",
      "name": "Debug bot",
      "program": "${workspaceFolder}/main.go",
      "env": {
        "DISCORD_TOKEN": "<your_token>",
        "GUILD_ID": "<your_guild_id>"
      }
    }
  ]
}
`
	err := os.Mkdir(".vscode", 0755)

	if err != nil {
		panic(err)
	}

	d := []byte(content)
	os.WriteFile(".vscode/launch.json", d, 0644)
}
