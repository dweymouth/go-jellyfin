# go-jellyfin

An API client library for Jellyfin music functionality for Go.

This code is an adaptation of the Jellyfin API implementation from the [Jellycli](https://github.com/tryffel/jellycli) project, originally written by Tero Vierimaa (@tryffel). Also a very helpful resource was the Jellyfin API layer of [Sonixd](https://github.com/jeffvli/sonixd).

The scope of this library is currently music-only, as it is being built for the [Supersonic](https://github.com/dweymouth/supersonic) project.

## Status

This library is functional, and currently in use in Supersonic. However, it is certainly not API stable and can be expected to change frequently.


## Example

```go
import (
    "github.com/dweymouth/go-jellyfin"
    "log"
    "context"
)

func main() {
    // create client
    jellyClient, err := jellyfin.NewClient("https://jellyfin.example.com", "supersonic", "1")
    if err != nil {
        log.Fatalf("unable to create jellyfin client: %v", err)
    }

    // login. Saves the access key to Client for future calls.
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := jellyClient.Login(ctx, "user", "pass"); err != nil {
        log.Fatalf("unable to log in to jellyfin: %v", err)
    }

    // get albums between 2000-2010
    filter := jellyfin.QueryOpts{
        Filter: jellyfin.Filter{
            YearRange: []int{2000, 2010},
        },
    }

    albums, err := jellyClient.GetAlbums(filter)
    if err != nil {
        log.Fatalf("unable to get albums: %v", err)
    }

    // print out all the album names
    for _, album := range albums {
        log.Print(album.Name)
    }
}

```