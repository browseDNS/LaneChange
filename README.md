## LaneChange
Receive an incoming HTTP request on `/` and delegate the response according to a previously indicated preference for the incoming IP.

If we don't have any LaneChange preferences in memory for the given IP, the request will receive the default response. The intent is for this endpoint to be hit many times, and have the response vary without needing to store any cookie or client information besides the IP.

The use case for this program is in network configurations where a user's control over their connected device is restrained by the device manufacturer. For instance, on the Nintendo Switch, a [captive portal](https://en.wikipedia.org/wiki/Captive_portal) check can completely prevent web access, and there's [nothing the user can do](https://www.change.org/p/nintendo-nintendo-expose-the-fully-functional-internet-browser-built-into-the-switch) about it.

### View Registered Lanes
Lanes can be configured in the config.json file, and are keyed by their name. To view all available lanes, issue a `GET /config`. This should match the json config file on the server, but is available for the client.

Example config and response:
```
{
    "default": "switchbru",
    "lanes": {
        "switchbru": {
            "headers": {
                "Content-Type": "text/html"
            },
            "content": "<script>location.href = 'https://dns.switchbru.com';</script>"
        },
        "nintendo": {
            "headers": {
                "Content-Type": "text/plain",
                "X-Organization": "Nintendo"
            },
            "content": "ok"
        }
    }
}
```

### Submitting LaneChange Preference
To submit a LaneChange preference for your given IP, make the call to `POST /change` with a JSON payload indicating which lane/endpoint name/key you'd like to switch to for your IP. You can also specify a duration in seconds to keep the preference before resetting to default. If no duration is specified, the preference will stay for the given IP until it is dropped (never expires).

Example LaneChange payload:
```
{
    "lane": "nintendo",
    "expires": "2020-10-24T01:45:53.570211-04:00"
}
```

To remove a LaneChange preference, issue a `DELETE /change` and any preference for your IP will be dropped, and you will receive the default lane for your next `GET /` request. Any subsequent `POST /change` request will also override a previous configuraiton.

To view your current preference, issue a `GET /change`. A 404 response on this endpoint indicates you have not yet made a preference, or it has been dropped for your IP.

### Compiling and Running
```
go get github.com/patrickmn/go-cache
go build
./LaneChange
```

### License
This program is licensed under the [GNU AGPLv3](https://choosealicense.com/licenses/agpl-3.0/).
