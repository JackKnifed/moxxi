Moxxi
=====

This is a HTTP proxy to allow you to access a site at a specific IP address via an alternate domain name.

The intention is to enable easier migrations.

Please see [the setup instructions](/setup.md) for setup information.

Design
------

This system consists of a few things:

* `nginx` runs on 443 and 80, and has a lot of server blocks to listen as
* `moxxi` - a custom written binary - listens on 8080 and generates configs with given input from each request - putting all those configs in one directory.
* `cron` deletes all configs older than a threshold in that directory.
* `syncthing` keeps that one directory in sync across the different servers in the cluster.
* A loadbalancer sits in front of it all, splitting traffic among the servers and providing redudancy.

`nginx` has two static server blocks:

* A default server block that simply serves one static file, a notification that the domain requested is expired, not configured, or broken.
* A proxy for contorl - requests htiting this server block are forwarded to `moxxi`.

`nginx` then also includes all the confs in a given folder, and is set to reload (reparse all configs) every 5 minutes.

`iptables` with incoming filtering is used to filter incoming traffic - only the SSH port, 80, and 443 are open. No filtering is done outgoing as this can hit any remote port on a remote server.

Layout
------

See below for a (very crude) visual diagram of how this works.

```

                               ----------------
                              | Load  Balancer |
          /-----------------  |  Port 80/443   | ---------------------\
          |                   |  *.domain.com  |                      |    
          |                    ----------------                       |      
          |                            |                              |
          v                            v                              v
                                                                                                              
 ----------------------       ----------------------      ---------------------- 
|       moxxi1         |     |       moxxi2         |    |       moxxi3         |
|                      |     |                      |    |                      |
| cleanup runs at 1am  |     | cleanup runs at 2am  |    | cleanup runs at 3am  |
|                      |     |                      |    |                      |
|   syncthing syncs    |     |   syncthing syncs    |    |   syncthing syncs    |
| /home/moxxi/vhosts.d |     | /home/moxxi/vhosts.d |    | /home/moxxi/vhosts.d |
 ----------------------       ----------------------      ----------------------
        \     /                      \     /                      \     /
        |     |                      |     |                      |     |
        |     |                      |     |                      |     |
        |     \______________________/     \______________________/     |
        |                                                               |
        |                     syncthing                                 |
        \_______________________________________________________________/


```

Each server is set up as:

```

    ________________________________________________________________  
   /                                                                \        
  /         ___________________              _____________           \
  |        /                   \  default   /             \          |
  |        |                   | ---------  | static page |          |
--|---80--*|                   |   vhost    |  domain.com |          |
  |        |       nginx       |            \_____________/          |
--|--443--*|                   |                                     |
  |        |                   |___         __________________       |
  |        |                   |   \       /                  \      |
  |        \___________________/   | vhost |  proxy to moxxi  |      |
  |             |           |      \_______| moxxi.domain.com |      |
  |     perodic | reload    |              \__________________/      |
  |         __________      | folder                     |           |
  |        /          \     | of                         |           |    
  |        |   cron   |     | vhosts                     * 8080      |
  |        \__________/     |                        ________        |    
  |             |           |                       /         \      |
  |       clean | vhosts ____________ creates ______|  moxxi  |      |
  |       over  | month /            \       /      \_________/      |
  |             \_______|  vhosts.d  | <----- vhosts                 |
  |                     \____________/                               |     
  |           ___________    ^                      ________         |
  |          /           \   | keeps               /        \        |
--|--22000--*| syncthing |---| in                  |  sshd  |*--22---|--
  |          \___________/  sync                   \________/        |
  \                                                                  /
   \________________________________________________________________/
  
``` 

JSON Format
-----------

JSON requests should be laid out as follows:

```json
[
  {
    "host": "hostname",
    "ip": "serverIP",
    "port": 443,
    "tls": true,
    "blockedHeaders": [
      "X-Frame-Options",
      "Accept-Encoding"
    ]
  },
  {
    "host": "liquidweb.com",
    "ip": "67.43.15.214",
    "port": 443,
    "tls": true,
    "blockedHeaders": [
      "X-Frame-Options",
      "Accept-Encoding"
    ]
  },
  {
    "host": "deleteos.com",
    "ip": "deleteos.com",
    "port": 443,
    "tls": true,
    "blockedHeaders": [
      "X-Frame-Options",
      "Accept-Encoding"
    ]
  }
]
```