# ForkServer
ForkServer for ForkPlayer<br>**(For testing purposes only!)**

### Run options:
*program* **[command]** **[options]**<br>
- where **[command]** (_for **windows** only, for other OS it is ignored_) is one of:<br>
**install** - installs and runs it as service (service name: **ForkServer**)<br>
**treeview** - sets the link to your media files folder (_in progress_, for now do it manually)<br>
**uninstall** - stops and removes the service<br><br>
- where **[options]** is a set of:<br>
**-a** [IP]{:PORT} - sets the address to listen http request to (default is :8027)<br>
**-d** {wd_path} - sets the path to server's working dir (default is current working dir)<br>
**-u** {number} - sets the period of {number} hours to automatically check for updates, 0 - means no autoupdates (default is 24)<br>
**-s** - sets to skip verify TLS certificates (can be usefull for slim and old OSes and routers)<br>
**-t** - turns off timestamps in logs (useful for systemd logging, because it places timestamps)<br>
**-i** - turns on info logging (useful for debugging)

*server logs errors to STDERR, info to STDOUT (**for windows service, it logs to the files: errors.log and info.log**)

### HTTP API:
http://{ServerIP}:{ServerPort}**[/path]**
- where **[/paht]** can be one of:<br>
**/** - web ui (*in progress*)<br>
**/test** - used by ForkPlayer testing the server<br>
**/test.json** - stata in json-format<br>
**/parserlink?{parsing_command}** - used by ForkPlayer parsing<br>
**/proxy.m3u8?link={url_of_m3u8}&[header={request_header}...]** - "proxy" the m.3u8 list<br>
**/torrserve[?link={url_of_torrent_or_magnet}][&shuffle=true]** - shows the playlist of the torrent content<br>
**/treeview** - main playlist<br>
**/treeview/[path]** - media files playlist<br>
**/set?Torrserve=search&search={addr}** - sets Torrserve address<br>
**/restart** - restarts the server<br>
**/udpate** - check the updates and install them if exists<br>
**/{pic}.svg**- show the embeded icons<br>
**/{plugin}/** - starts the plugin<br>
