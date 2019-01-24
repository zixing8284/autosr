# autosr: Automated Schelduled Recordings

autosr tracks streamers you are interested in.
When they go live autosr records the stream.

An external program such as streamlink or livestreamer is required.

Currently only SHOWROOM is supported.
Contributors would be greatly appreciated!

## Installing on Linux or OS X

autosr is a command-line application.
The following commands should be types into a terminal.

Download the right version from the [release page](https://github.com/bobbytrapz/autosr/releases/tag/v1.0.0-beta)

Install streamlink and autosr

```
sudo pip install streamlink
sudo mv autosr /usr/local/bin/autosr
# on OS X
sudo mv autosr-osx /usr/local/bin/autosr
```

## Installing on Windows

You can download the binary from the [release page](https://github.com/bobbytrapz/autosr/releases/tag/v1.0.0-beta).
Further instructions will be provided soon.

## Track streamers

Add streamers to track

```
autosr track
```

A text editor will open.
Put the url of the streamer you are interested in one line at a time.
The changes are applied without restarting.
For example,

```
https://www.showroom-live.com/MY_FAVORITE_STREAMER
```

## Start recording

Simply run:

```
autosr
```

You should see the autosr dashboard.
To exit press 'q'.

Even if you exit autosr is running in the background.

To stop it run:

```
autosr stop
```

## Watching videos

If anything is recorded, by default they can be found in your home directory in a 'autosr' directory.

The files will be in ts format. You can encode them however you like or just watch them.
I recommend using [mpv](https://mpv.io)

## Set custom options

The default options should be fine.
If you want to change them. For example, to use a different stream ripper or change the highlight colors:

```
autosr options
```

A configuration file should open for you to edit.
The changes are applied without restarting.

