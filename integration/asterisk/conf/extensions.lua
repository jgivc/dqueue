extensions = {}

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