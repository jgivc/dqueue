extensions = {}

extensions.public = {
    ["s"] = function(ctx, ext)
        app.hangup()
    end;
}

extensions.kamailio = {

    ["3000"] = function(ctx, ext)
        app["goto"]("queue-test", "s", 1)
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

extensions["queue-test"] = {
    s = function(ctx, ext)
        app.noop("Test queue")
        -- app.Set('CHANNEL(hangup_handler_push)=ivr-test-queue,s-HANGUP,1');
        -- uevent("stage=menu location=queue")
        app.agi("agi:async")
        app.hangup()
    end;

    ["s-CONNECT"] = function(ctx, ext)
        local ch = channel.CLIENT_CHANNEL:get()
        if ch ~= nil and ch ~= "" then
            -- uevent("stage=operator channel=" .. ch .. " operator=" ..channel.OPERATOR_NUMBER:get(), channel.CLIENT_UNIQUEID:get())
            -- app.bridge(ch, "F(ivr-test-queue^s-RATE^1)")
            local op_number = channel.OPERATOR_NUMBER:get()
            if op_number ~= nil then
                channel["CALLERID(num)"]:set(op_number)
            end
            app.bridge(ch)
        end
        app.noop("After bridge")
        app.hangup()
    end;

    --[[

    ["s-RATE"] = function(ctx, ext)
        uevent("stage=rate")
        app.Playback("custom/ivrhelp/rate_operator")
        app.WaitDigit(5,"12345")
        if channel.WAITDIGITSTATUS:get() == "DTMF" then
            uevent("stage=rate score=" ..channel.WAITDIGITRESULT:get())
            app.Playback("custom/ivrhelp/thank_you_for_rating")
            app.Playback("custom/ivrhelp/goodbye")
        end
        app.hangup()
    end;

    ["s-CANT-HANDLE"] = function(ctx, ext)
        uevent("stage=cannot_handle")
        app.Playback("custom/ivrhelp/cannot_handle")
        app.Playback("custom/ivrhelp/goodbye")
        app.hangup()
    end;

    ["s-HANGUP"] = function(ctx, ext)
        uevent("stage=hangup")
    end;

    ]]--
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