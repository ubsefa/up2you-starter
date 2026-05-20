const state = {
  tasks: [],
};

const els = {
  baseUrl: document.querySelector('#baseUrl'),
  tenantId: document.querySelector('#tenantId'),
  form: document.querySelector('#taskForm'),
  title: document.querySelector('#title'),
  description: document.querySelector('#description'),
  priority: document.querySelector('#priority'),
  dueDate: document.querySelector('#dueDate'),
  refreshBtn: document.querySelector('#refreshBtn'),
  publicBtn: document.querySelector('#publicBtn'),
  status: document.querySelector('#status'),
  tasks: document.querySelector('#tasks'),
  count: document.querySelector('#count'),
  publicList: document.querySelector('#publicList'),
};

function apiBase() {
  return els.baseUrl.value.replace(/\/+$/, '');
}

function headers(extra = {}) {
  return {
    'Content-Type': 'application/json',
    'X-Tenant-ID': els.tenantId.value.trim() || '00000000-0000-0000-0000-000000000001',
    'X-App-Key': 'my-todo',
    ...extra,
  };
}

function setStatus(message, isError = false) {
  els.status.textContent = message;
  els.status.classList.toggle('error', isError);
}

function unwrapList(payload) {
  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload?.data)) return payload.data;
  if (Array.isArray(payload?.data?.items)) return payload.data.items;
  if (Array.isArray(payload?.items)) return payload.items;
  if (Array.isArray(payload?.records)) return payload.records;
  return [];
}

function getId(task) {
  return task.id || task.entity_id || task.entityId;
}

function getState(task) {
  return task.current_state || task.state || 'open';
}

function escapeHtml(value) {
  return String(value ?? '').replace(/[&<>"']/g, char => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;',
  })[char]);
}

function toIsoDateTime(value) {
  return value ? new Date(value).toISOString() : undefined;
}

async function request(path, options = {}) {
  const response = await fetch(`${apiBase()}${path}`, {
    ...options,
    headers: headers(options.headers || {}),
  });
  const contentType = response.headers.get('content-type') || '';
  const payload = contentType.includes('application/json') ? await response.json() : await response.text();

  if (!response.ok || payload?.success === false) {
    const detail = payload?.error?.message || payload?.message || response.statusText;
    throw new Error(detail);
  }

  return payload;
}

async function loadTasks() {
  setStatus('Loading tasks...');
  try {
    const payload = await request('/api/v1/Task');
    state.tasks = unwrapList(payload);
    renderTasks();
    setStatus('Ready.');
  } catch (error) {
    setStatus(error.message, true);
  }
}

async function createTask(event) {
  event.preventDefault();
  const body = {
    title: els.title.value.trim(),
    description: els.description.value.trim(),
    priority: els.priority.value,
  };

  const dueDate = toIsoDateTime(els.dueDate.value);
  if (dueDate) body.due_date = dueDate;

  try {
    setStatus('Creating task...');
    await request('/api/v1/Task', {
      method: 'POST',
      body: JSON.stringify(body),
    });
    els.form.reset();
    els.priority.value = 'medium';
    await loadTasks();
  } catch (error) {
    setStatus(error.message, true);
  }
}

async function runTransition(task, transition) {
  const id = getId(task);
  if (!id) {
    setStatus('Task id is missing in the API response.', true);
    return;
  }

  const body = transition === 'archive' ? { reason: 'Archived from starter demo' } : {};

  try {
    setStatus(`Running ${transition}...`);
    await request(`/api/v1/Task/${encodeURIComponent(id)}/transitions/${transition}`, {
      method: 'POST',
      body: JSON.stringify(body),
    });
    await loadTasks();
  } catch (error) {
    setStatus(error.message, true);
  }
}

async function loadPublicTasks() {
  els.publicList.textContent = 'Loading public tasks...';
  els.publicList.classList.add('empty');

  try {
    const payload = await request('/api/v1/_public/public_open_tasks');
    const tasks = unwrapList(payload);
    renderPublicTasks(tasks);
  } catch (error) {
    els.publicList.textContent = error.message;
    els.publicList.classList.add('empty');
  }
}

function actionsFor(task) {
  const current = getState(task);
  const actions = [];

  if (current === 'open') actions.push(['start', 'Start'], ['archive', 'Archive']);
  if (current === 'in_progress') actions.push(['complete', 'Complete']);
  if (current === 'done') actions.push(['reopen', 'Reopen'], ['archive', 'Archive']);

  return actions;
}

function renderTasks() {
  els.count.textContent = String(state.tasks.length);

  if (!state.tasks.length) {
    els.tasks.innerHTML = '<p class="empty">No tasks yet.</p>';
    return;
  }

  els.tasks.innerHTML = state.tasks.map((task, index) => {
    const current = getState(task);
    const actions = actionsFor(task).map(([name, label]) => {
      const className = name === 'archive' ? 'danger' : 'secondary';
      return `<button class="${className}" type="button" data-index="${index}" data-action="${name}">${label}</button>`;
    }).join('');

    return `
      <article class="task">
        <div>
          <h3>${escapeHtml(task.title || 'Untitled task')}</h3>
          <p>${escapeHtml(task.description || 'No description.')}</p>
          <div class="meta">
            <span class="pill">${escapeHtml(current)}</span>
            <span class="pill">${escapeHtml(task.priority || 'medium')}</span>
            ${task.due_date ? `<span class="pill">${escapeHtml(new Date(task.due_date).toLocaleString())}</span>` : ''}
          </div>
        </div>
        <div class="actions">${actions || '<span class="pill">No actions</span>'}</div>
      </article>
    `;
  }).join('');
}

function renderPublicTasks(tasks) {
  if (!tasks.length) {
    els.publicList.textContent = 'No open public tasks.';
    els.publicList.classList.add('empty');
    return;
  }

  els.publicList.classList.remove('empty');
  els.publicList.innerHTML = tasks.map(task => (
    `<div class="compact-item">${escapeHtml(task.title || 'Untitled task')}</div>`
  )).join('');
}

els.form.addEventListener('submit', createTask);
els.refreshBtn.addEventListener('click', loadTasks);
els.publicBtn.addEventListener('click', loadPublicTasks);
els.tasks.addEventListener('click', event => {
  const button = event.target.closest('button[data-action]');
  if (!button) return;

  const task = state.tasks[Number(button.dataset.index)];
  if (task) runTransition(task, button.dataset.action);
});

loadTasks();
