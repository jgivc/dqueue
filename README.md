# Dqueue
[![Go Report Card](https://goreportcard.com/badge/github.com/jgivc/dqueue)](https://goreportcard.com/report/github.com/jgivc/dqueue)


Dqueue is a distributed queue between multiple asterisk servers when they are behind a sip proxy server such as kamailio.
Calls are managed by ami. Clients are queued through the agi(agi:async) dialplan application in the specified context.
The operator is called by ami action Originate. After the operator response, the channel enters the dqueue
via an agi:async dialplan application with a client id argument. The client ID can be obtained from the CLIENT_ID variable
of operator channel. After that, the channels will be combined into a bridge. Consider that application as mvp

## Usage

Config example
```yaml
context: dqueue                   #Context from which call app.agi(agi:async) to insert client to queue
ami:
  servers:
    - host: 10.0.0.101
      username: admin
      secret: password
    - host: 10.0.0.102
      username: admin
      secret: password
voip:
  tech_data: PJSIP/%s@kamailio    # ami action Originate Channel
  context: dqueue                 # ami action Originate Context
  exten: s-OPERATOR               # ami action Originate Exten
  dial_timeout: 20s
operator:
  operators:
    - 2001
    - 2002
```

Asterisk extension.lua example
```lua
extensions = {}

extensions.public = {
    ["s"] = function(ctx, ext)
        app.hangup()
    end;
}

extensions.kamailio = {

    ["3000"] = function(ctx, ext)
        app["goto"]("dqueue", "s", 1)
        app.hangup()
    end;

    include = { "LOCAL" }
}

extensions.LOCAL = {
    ["_[12]XXX"] = function(ctx, ext)
      local domain = channel.SIPDOMAIN:get()
        if domain == nil then
            app.Hangup(5)
            return
        end
        app.dial("PJSIP/kamailio/sip:" .. ext .. "@" .. domain .. ",60")
        app.Hangup()
    end;
}

extensions["dqueue"] = {
    s = function(ctx, ext)
        app.noop("New client " .. channel.CALLERID("num"):get())
        app.agi("agi:async")
        app.hangup()
    end;

    ["s-OPERATOR"] = function(ctx, ext)
        local client_id = channel.CLIENT_ID:get()
        app.noop("Operator " .. channel.CALLERID("num"):get() .. " for client: " .. client_id)
        app.agi("agi:async", client_id)
        app.Hangup()
    end;
}
```


Operators can be defined in config file:

```yaml
operator:
  operators:
    - 1001
    - 1002
```
or get by http:

```yaml
operator:
  api_url: https://...
  api_timeout: 10s
  no_verify: true
```
as json:

```json
{ "operators": [
  { "number": 1001,
    "last_name": "Operator",
    "first_name": "One"
  },
  { "number": 1002,
    "last_name": "Operator",
    "first_name": "Two"
  },
]}
```

You can view example setup in integration folder. And can run integration test with make e2e command (required docker and docker compose).

