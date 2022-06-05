# Abandoned: Sorry I don't use a setup that needs this anymore, if anyone want's to start a fork let me know I'll link it here

# Sonarr Torrent Cleaner

Simple executable to remove torrents and optionally blacklist them from Sonarr if they haven't progressed in a set time period

## Usage

- Run once to generate a default `config.yaml`
- Fill out the config with your Sonarr details
- Run the executable `SonarrTorrentCleaner(.exe)`
- Use something like screen to allow it to continue running perpetually (service installer coming up)

## Config

``` yaml
{
    # How often to poll sonarr for changes
    CheckTimeMinutes: 10
    # Time to wait on a torrent that has made no progress before removing it 
    NoProgressTimeoutMinutes: 30
    # The address your Sonarr install can be found at
    SonarrURL: http://localhost:8989
    # Your Sonarr api key
    SonarrAPIKey: "XXXXXXXXXXXXXXXXXXX"
    # Blacklist the torrent in Sonarr so it's not downloaded again
    Blacklist: true
}
```

## Run in screen

`screen -S SonarrTorrentCleaner`
`./SonarrTorrentCleaner`
`ctrl`+`shift`+`a`+`d`

## TODO

- Auto releases
- Snap package
- Docker image
