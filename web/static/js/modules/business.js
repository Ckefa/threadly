let currentClientId = null;

function showNewClientModal() {
  document.getElementById('new-client-modal').classList.remove('hidden');
}

function hideNewClientModal() {
  document.getElementById('new-client-modal').classList.add('hidden');
  document.getElementById('new-client-form').reset();
}

function openQuotationModal(clientId) {
  showNotification('Create quotation feature coming soon!', 'info');
}

function openPaymentModal(clientId) {
  showNotification('Payment request feature coming soon!', 'info');
}

function openGoalModal(clientId) {
  showNotification('Goal setting feature coming soon!', 'info');
}

function loadClient(clientId) {
  currentClientId = clientId;
  document.querySelectorAll('.client-item').forEach(item => {
    item.classList.remove('bg-teal-50', 'border-l-4', 'border-teal-600');
  });
  const el = document.querySelector(`[data-client-id="${clientId}"]`);
  if (el) el.classList.add('bg-teal-50', 'border-l-4', 'border-teal-600');
  htmx.ajax('GET', `clients/${clientId}/messages`, {
    target: '#chat-area',
    swap: 'innerHTML'
  });
}

function deleteClient(clientId, clientName) {
  if (!confirm(`Are you sure you want to delete "${clientName}"? This action cannot be undone.`)) return;
  fetch(`clients/${clientId}`, { method: 'DELETE', headers: { 'Content-Type': 'application/json' } })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Customer deleted successfully!', 'success');
        const el = document.querySelector(`[data-client-id="${clientId}"]`);
        if (el) el.remove();
        if (currentClientId == clientId) {
          document.getElementById('chat-area').innerHTML =
            `<div class="flex-1 flex items-center justify-center text-gray-400">
              <div class="text-center text-gray-500">
                <i class="fas fa-comments text-6xl mb-4"></i>
                <p>Select a client to start chatting</p>
              </div>
            </div>`;
          currentClientId = null;
        }
      } else {
        showNotification(data.error || 'Failed to delete client', 'error');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      showNotification('Failed to delete client', 'error');
    });
}

function toggleActionButtons() {
  const buttons = document.getElementById('action-buttons');
  const icon = document.getElementById('action-icon');
  const mainBtn = document.getElementById('main-action-btn');
  if (buttons.classList.contains('hidden')) {
    buttons.classList.remove('hidden');
    icon.classList.replace('fa-plus', 'fa-times');
    if (mainBtn) mainBtn.classList.add('rotate-45');
    buttons.querySelectorAll('button').forEach((btn, i) => {
      btn.style.cssText = 'opacity:0;transform:translateY(20px)';
      setTimeout(() => { btn.style.cssText = 'opacity:1;transform:translateY(0)'; }, i * 50);
    });
  } else {
    buttons.classList.add('hidden');
    icon.classList.replace('fa-times', 'fa-plus');
    if (mainBtn) mainBtn.classList.remove('rotate-45');
  }
}

function sendMessage() {
  const form = document.getElementById('message-form');
  if (form) form.submit();
}

function createMessageAction(messageId, type) {
  const title = prompt(`Create ${type}:`);
  if (!title) return;
  const description = prompt('Description (optional):') || '';
  const dueDate = type === 'booking' ? prompt('Date (YYYY-MM-DD):') : null;
  const fd = new FormData();
  fd.append('type', type);
  fd.append('title', title);
  fd.append('description', description);
  if (dueDate) fd.append('due_date', dueDate);
  htmx.ajax('POST', `/messages/${messageId}/actions`, {
    target: '#actions-panel', swap: 'innerHTML', values: fd
  });
}

function saveConversationProgress(clientId, stage) {
  fetch('clients/' + clientId + '/conversation-id')
    .then(r => r.json())
    .then(data => {
      if (!data.conversation_id) {
        showNotification('Failed to get conversation ID', 'error');
        return;
      }
      const fd = new FormData();
      fd.append('current_stage', stage);
      fd.append('progress_score', getProgressScore(stage));
      fetch('/conversations/' + data.conversation_id + '/stage', { method: 'PUT', body: fd })
        .then(r => r.ok ? showNotification('Conversation progress updated successfully!', 'success')
          : showNotification('Failed to update progress', 'error'))
        .catch(e => { console.error(e); showNotification('Failed to save conversation progress', 'error'); });
    })
    .catch(e => { console.error(e); showNotification('Failed to get conversation information', 'error'); });
}

document.addEventListener('DOMContentLoaded', function () {
  document.getElementById('new-client-form')?.addEventListener('submit', function (e) {
    e.preventDefault();
    fetch('clients', { method: 'POST', body: new FormData(this) })
      .then(r => r.json())
      .then(data => {
        if (data.success) {
          hideNewClientModal();
          showNotification('Client added successfully!', 'success');
          setTimeout(() => window.location.href = '/', 1500);
        } else {
          showNotification(data.error || 'Failed to add client', 'error');
        }
      })
      .catch(e => { console.error(e); showNotification('Failed to add client', 'error'); });
  });

  document.addEventListener('click', function (e) {
    const saveBtn = e.target.closest('.save-progress-btn');
    if (saveBtn) {
      const id = saveBtn.getAttribute('data-customer-id');
      const dd = document.querySelector(`.conversation-progress-dropdown[data-customer-id="${id}"]`);
      if (dd && dd.value) saveConversationProgress(id, dd.value);
    }
    const item = e.target.closest('.client-item');
    if (item && !e.target.closest('.conversation-progress-dropdown') && !e.target.closest('.save-progress-btn')) {
      loadClient(item.getAttribute('data-client-id'));
    }
  });
});
