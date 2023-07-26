extensions = {}

extensions.public = {
    ["s"] = function(ctx, ext)
        app.hangup()
    end;
}

extensions.kamailio = {

    ["3000"] = function(ctx, ext)
        app.agi("agi:async")
        app.hangup()
    end;

    ["4000"] = function(ctx, ext)
        app.answer()
        app.playback("beep")
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

        local callid = channel["PJSIP_HEADER(read,Call-ID)"]:get()
        if callid ~= nil then
            app.noop("Call-ID: " .. callid)
            channel.__CALLID:set(callid)
        end
        -- channel.__SIPDOMAIN:set(domain)
        -- sip_dial("kamailio/sip:" .. ext .. "@" .. domain)
        -- app.dial("PJSIP/kamailio/sip:" .. ext .. "@" .. domain .. ",60,Tt")
        app.dial("PJSIP/kamailio/sip:" .. ext .. "@" .. domain .. ",60,Ttb(DIALOUT^outhandler^1)")
        -- app.dial("PJSIP/" .. ext .."@kamailio,60,Tt")
        -- app.dial("PJSIP/kamailio/sip:" .. ext .. "@" .. domain .. ",60,Tt")
        app.noop("After Dial")
        app.Hangup()
    end;
}


extensions.DIALOUT = {
    ["outhandler"] = function(ctx, ext)
        
        local callid = channel.CALLID:get()
        if callid ~= nil then
            app.noop("X-Call-ID: " .. callid)
            -- channel.PJSIP_HEADER("add","X-Call-ID"):set(callid)
            channel.PJSIP_HEADER("add","X-Parent-Call-ID"):set(callid)
            -- channel.PJSIP_HEADER("add","X-Call-ID"):set(callid)
        end
        app.Return()
    end;
}


--[[
extensions.default = {

    ["_1XXX"] = function(ctx, ext)
        app.dial("PJSIP/" .. ext, 60)
        app.hangup()
    end;

    ["2000"] = function(ctx, ext)
        app.echo()
        app.hangup()
    end;

    ["3000"] = function(ctx, ext)
        app.agi("agi:async")
        app.hangup()
    end;

    ["4000"] = function(ctx, ext)
        app.answer()
        app.playback("beep")
        app.hangup()
    end;
}
]]--