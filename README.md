# Table Top Simulator Exporter

This tool is a program to export all your beloved games from within the tabletop simulator into your custom server.

This tool downloads all the resource necessary to run the game and replace all the domain provided by a third-party domain with a custom one provided by you.
This way you can self-host your own resources and avoid some of them getting lost once the legit holder of the right decides to delete them.

This tool is for backup purposes only, and it's provided as is.

## Usage
First of all, create some folders

```
mkdir output
mkdir logs
mkdir zip
```

Now you can run the exporter
the exporter creates a separate game folder inside of `output` folder.
each folder represents a game with all of its assets.
it also provides the `file_name.json`` that contains all the instructions for tts to run the game patched with a special token.

```
go run index.go
```

the exporter creates a separate game folder inside of `output` folder.
each folder represents a game with all of its assets.
it also provide the `file_name.json` that contains all the instruction for tts to run the game patched with a special token.

all the `file_name.json` got patched with `##EXPORTED_DOMAIN_NAME##/` so if you want to restore the game first of all you have to manually replace 
all the instances of `##EXPORTED_DOMAIN_NAME##/` with the custom domain folder where the assets are stored like:
`##EXPORTED_DOMAIN_NAME##/` -> `https://www.mypersonaldomain.com/tss_backup/game/`

## Manually restore

```
sed -i 's/##EXPORTED_DOMAIN_NAME##/https:\/\/www.mypersonaldomain.com\/tss_backup\/game/g' file_name.json
```

once replaced you should copy the file in the correct location

```
cp file_name.json ~/.local/share/Tabletop\ Simulator/Mods/Workshop/
```

once copied in the proper position you should also add a row into the `WorkshopFileInfos.json`

```
nano ~/.local/share/Tabletop Simulator/Mods/Workshop/WorkshopFileInfos.json
```

```
{
    "Directory": "/home/{user}//.local//share//Tabletop Simulator//Mods//Workshop/file_name.json",
    "Name": "A descriptive name",
    "UpdateTime": 1670630300
}
```

The program can be run multiple times since skips already downloaded folders.
You could check the outcome by looking into the detailed logs of the game.
Some assets may be lost forever and impossible to back up.
I suggest you test incomplete backups to be sure they work.

