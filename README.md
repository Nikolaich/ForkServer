# ForkServer
ForkServer for ForkPlayer<br>**(For testing purposes only!)**

### Run options:
*program* **[options]**<br>
Where **[options]** is a any set of:<br>
**-a** [IP]{:PORT} - sets the address to listen http request to (default is :8027)<br>
**-d** {wd_path} - sets the path to server's working dir (default is current working dir)<br>
**-u** {number} - sets the period of {number} hours to automatically check for updates, 0 - means no autoupdates (default is 24)<br>
**-n** {service_name} - sets the service name for the server (if it is not **ForkServer**) for **Windows only**, to restart it when updated<br>
**-s** - sets to skip verify TLS certificates (can be usefull for slim and old OSes and routers)<br>
**-t** - turns off timestamps in logs (useful for systemd logging, because it place timestamps)<br>
**-i** - turns off info logging (to reduce logging)

*server logs errors to STDERR, info to STDOUT (**for windows service, it logs to the files: errors.log and info.log**)

### HTTP API:
http://{ServerIP}:{ServerPort}**[/path]**
- where **[/paht]** can be one of:<br>
**/** - web ui (*in progress*)<br>
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
**/{pic}.svg**- shows the embeded icons<br>
**/{plugin}/** - starts the plugin<br>
