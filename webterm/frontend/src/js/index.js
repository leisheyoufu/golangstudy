var Terminal = require('xterm').Terminal;
Terminal.applyAddon(require('xterm/lib/addons/fit'));

var terminalContainer = document.getElementById('term-box');
term = new Terminal({
    cursorBlink: true
});
term.open(terminalContainer);
term.fit();
var viewport = document.querySelector('.xterm-viewport');

function log() {
    console.log(
        term.cols,
        term.rows,
        viewport.style.lineHeight,
        viewport.style.height
    );
}

log();

term.writeln('this is demo for "refresh-viewport-height"!');
term.writeln('press any key!');
log();

var url = 'ws://' + window.location.host + window.location.pathname + 'term';
var ws = new WebSocket(url);

/*
if (document.readyState === 'complete' || document.readyState !== 'loading') {
    openWs();
} else {
    document.addEventListener('DOMContentLoaded', openWs);
}
*/

term.on('key', function(key, ev) {
    // let's change the line-height!
    terminalContainer.style.lineHeight = '20px';
    term.fit();
    term.write(key);
    ws.send(key);
    log();
});