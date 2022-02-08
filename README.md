![](img/banner.jpg)
# Automatic Update Launcher
A general purpose updater for updating (web) applications server when update folder exists

## What is this?

The Automatic update launcher only do one simple task, which is to upgrade your application to the latest version if it finds a later version of such application in the ```./updates``` folder.

Its logic basically goes like this

1. If updates/ folder exists
   1. Backup the old application files and data
   2. Copy the updates for application from updates folder to current folder
   3. Remove the updates folder
2. Start the application
3. Check for crash, if crash happens after update
   1. Restore the old application files and data
   2. Remove the backup folder
4. Restart the whole process unless crash retry count > max allowed



## Build

Require Go 1.17 or above

```
go mod tidy
go build
```

or just grab a copy from the release page.

## Usage

Configure the launcher with launcher.json. See the example below

```
{
    "version":"1.0",
    "start":"helloworld*",
    "backup": ["./*.config","./*.exe"],
    "verbal": true,
    "resp_port": 25576,
    "max_retry": 3,
    "crash_time": 3
}
```

| Key        | Usage                                                        |
| ---------- | ------------------------------------------------------------ |
| Version    | Launcher version                                             |
| Start      | Start path, support wildcard, will start the first matching file |
| Backup     | Files to be backed up before updates, will put into app.old  |
| Verbal     | Show launcher verbal message, use for debugs                 |
| Resp_port  | The port for response check in application. See "Check Launcher" section for more information |
| Max_retry  | Max retry before restoring from old application backup       |
| Crash_time | The time difference for start to exit is less than this value, the application is assume crashed |

### Check Launcher in Application

If your application needs to check if the launcher exists, you can create an HTTP GET request to the ```http://localhost://{resp_port}/chk``` endpoint to see if there are any response from the launcher. The launcher will response with a text message matching the pattern ```Launcher v" + launchConfig.Version```

