let currentClientId = null;

function showNewClientModal() {
  document.getElementById('new-client-modal').classList.remove('hidden');
}

function hideNewClientModal() {
  document.getElementById('new-client-modal').classList.add('hidden');
  document.getElementById('new-client-form').reset();
}

function openQuickBookingModal(clientId) {
  populateClientData(clientId);
  loadServicesForBooking();
  document.getElementById('quickBookingModal').classList.remove('hidden');
}

function openQuickOrderModal(clientId) {
  populateClientData(clientId);
  loadProductsForOrder();
  document.getElementById('quickOrderModal').classList.remove('hidden');
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

function populateClientData(clientId) {
  const el = document.querySelector(`[data-client-id="${clientId}"]`);
  if (el) {
    const name = el.querySelector('h3')?.textContent || '';
    const email = el.querySelector('.text-gray-500')?.textContent || '';
    document.getElementById('orderCustomerName').value = name;
    document.getElementById('orderCustomerEmail').value = email;
    document.getElementById('bookingCustomerName').value = name;
    document.getElementById('bookingCustomerEmail').value = email;
  }
}

async function loadProductsForOrder() {
  try {
    const response = await fetch('/products');
    const data = await response.text();
    const parser = new DOMParser();
    const doc = parser.parseFromString(data, 'text/html');
    const productCards = doc.querySelectorAll('[data-product-id]');
    const select = document.getElementById('orderProduct');
    select.innerHTML = '<option value="">Choose a product...</option>';
    productCards.forEach(card => {
      const opt = document.createElement('option');
      opt.value = card.dataset.productId;
      opt.textContent = card.querySelector('h3')?.textContent || 'Unknown Product';
      select.appendChild(opt);
    });
  } catch (error) {
    console.error('Failed to load products:', error);
    showNotification('Failed to load products', 'error');
  }
}

async function loadServicesForBooking() {
  try {
    const response = await fetch('/services');
    const data = await response.text();
    const parser = new DOMParser();
    const doc = parser.parseFromString(data, 'text/html');
    const serviceCards = doc.querySelectorAll('[data-service-id]');
    const select = document.getElementById('bookingService');
    select.innerHTML = '<option value="">Choose a service...</option>';
    serviceCards.forEach(card => {
      const opt = document.createElement('option');
      opt.value = card.dataset.serviceId;
      opt.textContent = card.querySelector('h3')?.textContent || 'Unknown Service';
      select.appendChild(opt);
    });
  } catch (error) {
    console.error('Failed to load services:', error);
    showNotification('Failed to load services', 'error');
  }
}

function hideQuickOrderModal() {
  document.getElementById('quickOrderModal').classList.add('hidden');
  document.getElementById('quickOrderForm').reset();
}

function hideQuickBookingModal() {
  document.getElementById('quickBookingModal').classList.add('hidden');
  document.getElementById('quickBookingForm').reset();
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
  document.getElementById('quickOrderForm')?.addEventListener('submit', function (e) {
    e.preventDefault();
    const data = {
      client_id: parseInt(currentClientId),
      product_id: parseInt(document.getElementById('orderProduct').value),
      quantity: parseInt(document.getElementById('orderQuantity').value),
      customer_name: document.getElementById('orderCustomerName').value,
      customer_email: document.getElementById('orderCustomerEmail').value,
      customer_phone: document.getElementById('orderCustomerPhone').value,
      delivery_address: document.getElementById('orderDeliveryAddress').value,
      notes: document.getElementById('orderNotes').value
    };
    fetch('/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
      .then(r => r.json())
      .then(data => {
        if (data.success) {
          const sel = document.getElementById('orderProduct');
          const name = sel.options[sel.selectedIndex].text;
          hideQuickOrderModal();
          showNotification(data.message || 'Order created successfully!', 'success');
          if (data.order) addOrderMessageToChat({ ...data.order, product_name: name, quantity: data.quantity });
        } else {
          showNotification(data.error || 'Failed to create order', 'error');
        }
      })
      .catch(e => { console.error(e); showNotification('Failed to create order', 'error'); });
  });

  document.getElementById('quickBookingForm')?.addEventListener('submit', function (e) {
    e.preventDefault();
    const bookingDate = document.getElementById('bookingDate').value;
    const bookingTime = document.getElementById('bookingTime').value;
    const [hours, minutes] = bookingTime.split(':');
    const formattedTime = `${hours.padStart(2, '0')}:${minutes.padStart(2, '0')}`;
    const data = {
      client_id: parseInt(currentClientId),
      service_id: parseInt(document.getElementById('bookingService').value),
      customer_name: document.getElementById('bookingCustomerName').value,
      customer_email: document.getElementById('bookingCustomerEmail').value,
      customer_phone: document.getElementById('bookingCustomerPhone').value,
      booking_date: `${bookingDate}T${formattedTime}`,
      notes: document.getElementById('bookingNotes').value
    };
    fetch('/bookings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
      .then(r => r.json())
      .then(data => {
        if (data.success) {
          const sel = document.getElementById('bookingService');
          const name = sel.options[sel.selectedIndex].text;
          hideQuickBookingModal();
          showNotification(data.message || 'Booking created successfully!', 'success');
          if (data.booking) addBookingMessageToChat({ ...data.booking, service_name: name });
        } else {
          showNotification(data.error || 'Failed to create booking', 'error');
        }
      })
      .catch(e => { console.error(e); showNotification('Failed to create booking', 'error'); });
  });

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
