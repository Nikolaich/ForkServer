# ForkServer
ForkServer is a program for ForkPlayer support (ex. RemoteFork). General purposes are:
- preser support (ForkPlayer embeded functionality)
- view and play media files form your server/computer
- view and play torrent files (additional software needed, e.g. Torrserver) 
- plugins system to extend functionality like IPTV, movies from the internet, etc (can be writen by yourself).<br>
- See preinstalled plugins: https://github.com/damiva/ForkServerPlugs
## Usage
### Run as a program
Download the applicable file from https://github.com/damiva/ForkServer/Releases to a it's working directory and execute it.<br>
In web browser go to http://your.server.ip:8027 to check and read how to use it in ForkPlayer
### Run options
ForkServer **[-a [IP]{:PORT}] [-d {WD_PATH}] [-u {NUMBER}] [-n {SERVICE_NAME}] [-s] [-i]**<br>where:
- **-a [IP]{:PORT}** - sets the address to listen http request to (default is :8027)<br>
- **-d {WD_PATH}** - sets the path to server's working dir (default is current working dir)<br>
- **-u {NUMBER}** - sets the period of {number} hours to automatically check for updates, 0 - means no autoupdates (default is 24)<br>
- **-n {SERVICE_NAME}** - sets the service name for the server for **Windows only**, to restart it when updated (default is **ForkServer**)<br>
- **-s** - sets to skip verify TLS certificates (can be usefull for slim and old OSes and routers)<br>
- **-i** - turns off info logging (to reduce logging)

Server logs errors and warnings to STDERR, info to STDOUT (**as Windows service, it logs to standard Windows event log**)
### Install as a service
#### Windows
1. Create folder "**C:\ForkServer**"
2. Download file **ForkServer-windows-*** for your architecture to the folder
3. Rename downloaded file to **ForkServer.exe**
4. To register and run the service: open **cmd** as Administrator and run the following three commands:<br>
>`sc create ForkServer binpath= "C:\ForkServer\ForkServer.exe -i" start= auto DisplayName= ForkServer`<br>
>`sc description ForkServer "ForkServer for ForkPlayer"`<br>
>`net start ForkServer`
#### Linux with systemd
1. Create folder **/opt/ForkServer**<br>#`mkdir /opt/ForkServer`
2. Download file **ForkServer-linux-*** for your architecture to the folder under name **ForkServer**<br>
#`curl -o /opt/ForkServer/ForkServer https://github.com/damiva/ForkServer/releases/download/0.08/ForkServer-linux-*`
3. Download file **ForkServer.service** to the folder **/etc/systemd/system/**<br>
#`curl -o /etc/systemd/system/ForkServer.service https://github.com/damiva/ForkServer/releases/download/0.08/ForkServer.service`
4. Install and run the service<br>#`systemctl enable ForkServer`<br>#`dydtemctl start ForkServer`
#### Mac OS X
*in progress*
## HTTP API:
http://{ServerIP}:{ServerPort}**[/path]**<br>
where **[/paht]** can be one of:<br>
**/** - web ui<br>
**/test** - used by ForkPlayer testing the server<br>
**/test.json** - stata in json-format<br>
**/parserlink?{parsing_command}** - used by ForkPlayer parsing<br>
**/proxy.m3u8?link={url_of_m3u8}[&header={{Name}:{value}}...]** - "proxy" the m3u8 list<br>
**/torrserve[?link={url_of_torrent_or_magnet}[&shuffle=true]]** - shows the playlist of the torrent content<br>
**/treeview** - main playlist<br>
**/treeview/[path][?shuffle=true]** - media files playlist<br>
**/set?Torrserve=search&search={addr}** - sets Torrserve address<br>
**/restart** - restarts the server<br>
**/udpate** - checks the updates and install them if exists<br>
**/gc** - run garbage collector (memory cleaning)<br>
**/{pic}.svg**- shows the embeded icon<br>
**/{plugin}/** - starts the plugin<br>
