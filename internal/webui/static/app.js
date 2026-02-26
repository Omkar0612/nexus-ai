// NEXUS AI v1.6 — Web UI client
// Handles: SSE agent activity stream, streaming chat, UI state

const messagesEl = document.getElementById('messages');
const inputEl    = document.getElementById('msg-input');
const sendBtn    = document.getElementById('send-btn');
const statusBar  = document.getElementById('status-bar');

// ── SSE Agent Activity Stream ──────────────────────────────────────────
const evtSource = new EventSource('/api/events');

evtSource.addEventListener('ping', () => {
  statusBar.textContent = '● Connected to NEXUS agent stream';
  statusBar.style.color = '#4ade80';
});

evtSource.onmessage = (e) => {
  try {
    const evt = JSON.parse(e.data);
    const icon = evt.status === 'running' ? '⚡' : evt.status === 'error' ? '✗' : '✓';
    statusBar.textContent = `${icon} [${evt.agent || 'router'}] ${evt.status}${
      evt.message ? ': ' + evt.message.slice(0, 80) : ''
    }`;
    statusBar.style.color = evt.status === 'error' ? '#f87171' : '#a3a3a3';
  } catch (_) {}
};

evtSource.onerror = () => {
  statusBar.textContent = '○ Agent stream disconnected — retrying…';
  statusBar.style.color = '#f87171';
};

// ── Chat ───────────────────────────────────────────────────────────────
function appendMsg(role, text) {
  const div = document.createElement('div');
  div.className = `msg ${role}`;
  div.textContent = text;
  messagesEl.appendChild(div);
  messagesEl.scrollTop = messagesEl.scrollHeight;
  return div;
}

async function sendMessage() {
  const text = inputEl.value.trim();
  if (!text) return;

  inputEl.value = '';
  sendBtn.disabled = true;
  appendMsg('user', text);

  const aiDiv = appendMsg('ai', '');
  aiDiv.classList.add('dot-loader');

  try {
    const res = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: text }),
    });

    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    if (!res.body) throw new Error('No response body');

    aiDiv.classList.remove('dot-loader');
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      // Parse SSE chunks: "data: <content>\n\n"
      const parts = buffer.split('\n\n');
      buffer = parts.pop() ?? '';
      for (const part of parts) {
        const line = part.startsWith('data: ') ? part.slice(6) : part;
        if (line) {
          aiDiv.textContent += line;
          messagesEl.scrollTop = messagesEl.scrollHeight;
        }
      }
    }
  } catch (err) {
    aiDiv.classList.remove('dot-loader');
    aiDiv.textContent = `Error: ${err.message}`;
    aiDiv.style.borderColor = '#f87171';
  } finally {
    sendBtn.disabled = false;
    inputEl.focus();
  }
}

sendBtn.addEventListener('click', sendMessage);
inputEl.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
});
