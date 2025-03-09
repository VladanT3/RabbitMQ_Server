# Pub/Sub server for a 'Risk-like' game called 'Peril'

The code for game logic is prewritten. I wrote the backend logic that uses RabbitMQ to transfer messages between servers and clients.

## About

The project is divided in 3 distinct sides:
- Game logic
- Server
- Client(s)

The server handles things like pausing and resuming the game, or reading game logs with:
- pause - Pause the game.
- resume - Resume the game.
- quit - Close the server.
- help - Show all possible commands.

Clients can play the game and have acces to commands:
- spawn:

    Possible units to spawn are:
    - infantry
    - cavalry
    - artillert

    Possible locations where units can be spawned:
    - europe
    - asia
    - americas
    - antarctica
    - africa
    - australia
- move - Clients can move their units around the map by specifying the location and unit ID.
- status - Returns data on which units you have available and where.
- spam - A command used to spam the server with game logs. It's really just a testing feature.
- quit - Exit the game.

## Structure

cmd/ - contains the code for the server and clients.

internal/gamelogic/ - prewritten game logic.

internal/pubsub/ - everything RabbitMQ related besides connecting to it (that's in the server and client).

internal/routing/ - routing constants for exchange and queue names and keys.

## Running the 'Creature' (project)

```
./rabbit.sh [start|stop|logs]
```
Runs the RabbitMQ docker container.

```
go run ./cmd/server
```
Runs the server.
```
go run ./cmd/client
```
Runs the client.

No I didn't build binaries, I don't care this was faster when constantly testing code.

```
./multiserver.sh [n]
```
Runs 'n' RabbitMQ servers. Used when testing backpressure to clear out full queues. No real reason to use it otherwise.
