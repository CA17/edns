# setecs

Setecs is CoreDNS Plugin, Set EDNS according to client source location 


# Configuration

    setecs {
        ecs-binding 114.114.114.114 clients lips.conf
        ecs-binding 8.8.8.8 clients 172.21.66.137 192.168.0.1/24
        ecs-table ecs-tables.txt
        reload 10s
        debug
    }

## ecs-binding

Set ecs for multiple client sources

    ecs-binding <ecs addr> clients <client addr | file | url>...

Content format:

    172.21.1.16
    172.21.2.0/24


## ecs-table

Single address mapping
    
    ecs-table <addr file | url>...

Content format:

    172.21.1.16:ecsip
    172.21.2.0/24:ecsip