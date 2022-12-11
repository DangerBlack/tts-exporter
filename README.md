# Table Top Simulator Exporter

This tool is a program to export all your beloved games from within tabletop simulator into your custom server.

This tools download all the resource necessary to run the game and replace all the domain provided by third part domain to a custom one provided by you.
This way you can self host your own resources and avoid that some of them get lost once the legit holder of the right decide to delete them.

This tools is for backup purpose only, and its provided as is.

## Usage

First of all create some folders

```
mkdir output
mkdir logs
mkdir zip
```

Now you can run the exporter

```
go run index.go
```

the exporter create separate game folder inside of `output` folder.
each folder represent a game with all of its assets.
it also provide the `file_name.json` that contains all the instruction for tts to run the game patched with a special token.

all the `file_name.json` got patched with `##EXPORTED_DOMAIN_NAME##/` do if you want to restore the game first of all you have to manually replace 
all the instance of `##EXPORTED_DOMAIN_NAME##/` with the custom domain folder where the assets are stored like:
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

Program can be run multiple time since skip already downloaded folders.
You could check the outcome by looking into the detailed logs of the game.
Some assets may be lost forever and impossible to backup.
I suggest you to test incomplete backup to be sure if they works.

