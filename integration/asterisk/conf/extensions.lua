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

    ["4000"] = function(ctx, ext)
        app.answer()
        app.playback("beep")
        app.hangup()
    end;

    ["5000"] = function(ctx, ext)
        app.answer()
        app.echo()
        app.hangup()
    end;

    include = { "LOCAL" }
}

extensions.LOCAL = {
    ["_[12]XXX"] = function(ctx, ext)
        app.noop("### DIAL to: " .. ext)
        local domain = channel.SIPDOMAIN:get()
        if domain == nil then
            app.Hangup(5)
            return
        end

        local callid = channel["PJSIP_HEADER(read,Call-ID)"]:get()
        if callid ~= nil then
            app.noop("Call-ID: " .. callid)
            channel.__CALLID:set(callid)
        end
        app.dial("PJSIP/kamailio/sip:" .. ext .. "@" .. domain .. ",60")
        app.noop("After Dial")
        app.Hangup()
    end;
}

extensions["dqueue"] = {
    s = function(ctx, ext)
        app.noop("Test queue")
        app.agi("agi:async")
        app.hangup()
    end;

    ["s-OPERATOR"] = function(ctx, ext)
        app.agi("agi:async", channel.CLIENT_ID:get())
        app.Hangup()
    end;
}
