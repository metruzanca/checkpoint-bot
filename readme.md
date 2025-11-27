<img src=".github/Remote_Caffeinator.png" alt="Risk of Rain 2's remote caffeinator">

# Caffeinator (aka Checkpoint Bot)

Little accountability bot for my discord server.

## Structure

Main business logic is split between:

- `internal/server/commands/` Defining slash commands and handlers
- `internal/database/{queries,migrations}` Defining db actions

## TODO

- [x] Integrate Goose for migrations (`schema` folder currently not being used)
- [x] Implement slash commands
  - [x] Create Checkpoint
  - [x] Get next checkpoint
  - ... more to come
- [x] (try to) Make sure bot unregisters ~~commands when it crashes~~ when starting
- [x] create second bot for dev

## Resources

- [Awesome DiscordGo](https://github.com/bwmarrin/discordgo/wiki/Awesome-DiscordGo)
- [Discordgo Examples](https://github.com/bwmarrin/discordgo/tree/master/examples)
