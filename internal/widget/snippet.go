package widget

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// handleSnippet serves the embeddable widget JS. The customer pastes
// one <script> tag and gets a floating bubble that opens a chat panel.
//
// This is a single self-contained JS file — no build step, no CDN,
// zero external deps. Small enough to inline in production too.
func handleSnippet(c *gin.Context) {
	apiBase := config.Get().BaseURL()

	c.Header("Content-Type", "application/javascript; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=300")
	fmt.Fprintf(c.Writer, snippetJS, apiBase)
}

// snippetJS is a single %s placeholder for the API base URL. The rest
// is plain JS/CSS — a floating bubble that expands into a chat panel,
// posts via /widget/message, and listens on /widget/stream for replies.
const snippetJS = `(function() {
  var script = document.currentScript;
  var key = script && script.getAttribute('data-key');
  if (!key) { console.warn('TalkEx widget: missing data-key'); return; }
  var API = %q;
  var sessionKey = 'talkex_widget_session_' + key;
  var STORAGE = window.localStorage;

  // ---------- styles ----------
  var css = ` + "`" + `
    #tx-bubble{position:fixed;bottom:24px;right:24px;width:60px;height:60px;border-radius:50%;background:#2563eb;color:#fff;display:flex;align-items:center;justify-content:center;cursor:pointer;box-shadow:0 6px 20px rgba(0,0,0,.2);z-index:2147483646;font-size:26px;transition:transform .15s}
    #tx-bubble:hover{transform:scale(1.06)}
    #tx-panel{position:fixed;bottom:100px;right:24px;width:340px;max-width:calc(100vw - 32px);height:480px;max-height:calc(100vh - 140px);background:#fff;border-radius:14px;box-shadow:0 12px 40px rgba(0,0,0,.18);display:none;flex-direction:column;overflow:hidden;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;z-index:2147483647}
    #tx-panel.open{display:flex}
    #tx-header{background:#2563eb;color:#fff;padding:14px 16px;font-weight:600;font-size:15px;display:flex;align-items:center;justify-content:space-between}
    #tx-close{background:none;border:0;color:#fff;font-size:20px;cursor:pointer;line-height:1;padding:0 4px}
    #tx-msgs{flex:1;overflow-y:auto;padding:14px;background:#f9fafb;font-size:13px;color:#111}
    .tx-msg{max-width:80%;padding:8px 12px;border-radius:14px;margin-bottom:8px;line-height:1.35;word-wrap:break-word}
    .tx-msg.tx-out{background:#fff;border:1px solid #e5e7eb;color:#111;border-bottom-left-radius:4px}
    .tx-msg.tx-in{background:#2563eb;color:#fff;margin-left:auto;border-bottom-right-radius:4px}
    #tx-form{display:flex;padding:10px;border-top:1px solid #e5e7eb;background:#fff}
    #tx-input{flex:1;border:1px solid #d1d5db;border-radius:8px;padding:8px 10px;font-size:13px;outline:none;font-family:inherit}
    #tx-input:focus{border-color:#2563eb}
    #tx-send{margin-left:8px;background:#2563eb;color:#fff;border:0;border-radius:8px;padding:0 14px;font-weight:600;font-size:13px;cursor:pointer}
    #tx-send:disabled{opacity:.5;cursor:not-allowed}
    #tx-brand{padding:6px;text-align:center;font-size:10px;color:#9ca3af;background:#f9fafb;border-top:1px solid #f3f4f6}
    #tx-brand a{color:#6b7280;text-decoration:none}
  ` + "`" + `;
  var s = document.createElement('style'); s.textContent = css; document.head.appendChild(s);

  // ---------- DOM ----------
  var bubble = document.createElement('div');
  bubble.id = 'tx-bubble';
  bubble.innerHTML = '&#128172;';
  document.body.appendChild(bubble);

  var panel = document.createElement('div');
  panel.id = 'tx-panel';
  panel.innerHTML =
    '<div id="tx-header"><span id="tx-title">Chat</span><button id="tx-close">&times;</button></div>' +
    '<div id="tx-msgs"></div>' +
    '<form id="tx-form"><input id="tx-input" placeholder="Type a message..." autocomplete="off"/><button id="tx-send" type="submit">Send</button></form>' +
    '<div id="tx-brand">Powered by <a href="https://talkex.io" target="_blank">TalkEx</a></div>';
  document.body.appendChild(panel);

  var msgs = panel.querySelector('#tx-msgs');
  var input = panel.querySelector('#tx-input');
  var form = panel.querySelector('#tx-form');
  var title = panel.querySelector('#tx-title');
  var header = panel.querySelector('#tx-header');

  bubble.onclick = function() { panel.classList.add('open'); ensureSession(); input.focus(); };
  panel.querySelector('#tx-close').onclick = function() { panel.classList.remove('open'); };

  // ---------- session ----------
  function loadConfig() {
    return fetch(API + '/widget/config?key=' + encodeURIComponent(key))
      .then(function(r){ return r.ok ? r.json() : null; })
      .then(function(cfg){
        if (!cfg) return;
        title.textContent = cfg.title || 'Chat';
        bubble.style.background = cfg.theme_color || '#2563eb';
        header.style.background = cfg.theme_color || '#2563eb';
        var sendBtn = form.querySelector('#tx-send');
        if (sendBtn) sendBtn.style.background = cfg.theme_color || '#2563eb';
        if (cfg.greeting) addMsg(cfg.greeting, 'out');
      });
  }

  var sessionId = null;
  function ensureSession() {
    if (sessionId) return Promise.resolve();
    var stored = STORAGE.getItem(sessionKey);
    if (stored) {
      sessionId = stored;
      return loadConfig().then(loadHistory).then(subscribe);
    }
    return loadConfig().then(function() {
      return fetch(API + '/widget/init', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({key: key, page_url: window.location.href})
      });
    }).then(function(r){ return r.json(); }).then(function(data) {
      sessionId = data.session_id;
      STORAGE.setItem(sessionKey, sessionId);
      subscribe();
    }).catch(function(e){ console.error('tx init failed', e); });
  }

  function loadHistory() {
    return fetch(API + '/widget/messages?key=' + encodeURIComponent(key) + '&session_id=' + encodeURIComponent(sessionId))
      .then(function(r){ return r.ok ? r.json() : []; })
      .then(function(list){
        (list || []).forEach(function(m){
          if (m.body === '[chat started]') return; // hide the init marker
          addMsg(m.body, m.direction === 'outbound' ? 'out' : 'in');
        });
      })
      .catch(function(){});
  }

  function subscribe() {
    try {
      var es = new EventSource(API + '/widget/stream?key=' + encodeURIComponent(key) + '&session_id=' + encodeURIComponent(sessionId));
      es.addEventListener('message', function(e){
        try { var m = JSON.parse(e.data); addMsg(m.body, 'out'); } catch(_){}
      });
      es.onerror = function(){ es.close(); setTimeout(subscribe, 3000); };
    } catch(_){}
  }

  function addMsg(body, dir) {
    var div = document.createElement('div');
    div.className = 'tx-msg tx-' + dir;
    div.textContent = body;
    msgs.appendChild(div);
    msgs.scrollTop = msgs.scrollHeight;
  }

  form.onsubmit = function(e) {
    e.preventDefault();
    var body = input.value.trim();
    if (!body || !sessionId) return;
    input.value = '';
    addMsg(body, 'in');
    fetch(API + '/widget/message', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({key: key, session_id: sessionId, body: body})
    }).catch(function(e){ console.error('tx send failed', e); });
  };
})();
`
