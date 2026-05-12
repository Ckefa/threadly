

function openClientOrderModal() {
  loadClientProducts();
  document.getElementById('clientOrderModal').classList.remove('hidden');
}

function hideClientOrderModal() {
  document.getElementById('clientOrderModal').classList.add('hidden');
  document.getElementById('clientOrderForm').reset();
}

function openClientBookingModal() {
  loadClientServices();
  document.getElementById('clientBookingModal').classList.remove('hidden');
}

function hideClientBookingModal() {
  document.getElementById('clientBookingModal').classList.add('hidden');
  document.getElementById('clientBookingForm').reset();
}

async function loadClientProducts() {
  try {
    const response = await fetch(`/client/businesses/${businessId}/products`);
    const products = await response.json();
    const select = document.getElementById('clientOrderProduct');
    select.innerHTML = '<option value="">Choose a product...</option>';
    products.forEach(p => {
      const opt = document.createElement('option');
      opt.value = p.id;
      opt.textContent = `${p.name} - $${p.price}`;
      select.appendChild(opt);
    });
  } catch (error) {
    console.error('Error loading products:', error);
  }
}

async function loadClientServices() {
  try {
    const response = await fetch(`/client/businesses/${businessId}/services`);
    if (!response.ok) {
      if (response.status === 401) {
        showNotification('Please login to access services', 'error');
        return;
      }
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const services = await response.json();
    const select = document.getElementById('clientBookingService');
    select.innerHTML = '<option value="">Choose a service...</option>';
    if (services.length === 0) {
      select.innerHTML = '<option value="">No services available</option>';
      showNotification('No services available for booking', 'warning');
      return;
    }
    services.forEach(s => {
      const opt = document.createElement('option');
      opt.value = s.id;
      opt.textContent = `${s.name} - $${s.min_price || s.max_price || 'Price not set'}`;
      select.appendChild(opt);
    });
  } catch (error) {
    console.error('Error loading services:', error);
    showNotification('Failed to load services', 'error');
  }
}

function submitOrderForm() {
  const productSelect = document.getElementById('clientOrderProduct');
  const quantityInput = document.getElementById('clientOrderQuantity');
  const addressInput = document.getElementById('clientOrderAddress');
  const notesInput = document.getElementById('clientOrderNotes');

  if (!productSelect.value) return showNotification('Please select a product', 'error');
  if (!quantityInput.value || quantityInput.value < 1) return showNotification('Please enter a valid quantity', 'error');

  const data = {
    product_id: parseInt(productSelect.value),
    quantity: parseInt(quantityInput.value),
    delivery_address: addressInput.value,
    notes: notesInput.value,
    business_id: parseInt(businessId)
  };

  fetch('/client/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        hideClientOrderModal();
        if (data.order) {
          addOrderMessageToChat({ ...data.order, product_name: data.product_name, quantity: data.quantity });
        }
        showNotification('Order request sent successfully!', 'success');
      } else {
        showNotification(data.error || 'Failed to send order request', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to send order request', 'error'); });
}

function submitBookingForm() {
  const serviceSelect = document.getElementById('clientBookingService');
  const dateInput = document.getElementById('clientBookingDate');
  const timeInput = document.getElementById('clientBookingTime');
  const notesInput = document.getElementById('clientBookingNotes');

  if (!serviceSelect.value) return showNotification('Please select a service', 'error');
  if (!dateInput.value) return showNotification('Please select a date', 'error');
  if (!timeInput.value) return showNotification('Please select a time', 'error');

  const bookingDateTime = `${dateInput.value}T${timeInput.value}:00Z`;

  fetch('/client/bookings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      service_id: parseInt(serviceSelect.value),
      scheduled_date: bookingDateTime,
      notes: notesInput.value,
      business_id: businessId
    })
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        hideClientBookingModal();
        if (data.booking) {
          addBookingMessageToChat({ ...data.booking, service_name: data.service_name });
        }
        showNotification('Booking request sent successfully!', 'success');
      } else {
        showNotification(data.error || 'Failed to send booking request', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to send booking request', 'error'); });
}

scrollToBottom();
markAsRead();
startMessagePolling();

function scrollToBottom() {
  var container = document.getElementById('messages-container');
  if (container) requestAnimationFrame(function () {
    container.scrollTop = container.scrollHeight;
  });
}

let pollingInterval = null;

function markAsRead() {
  fetch(`/client/businesses/${businessId}/read`, { method: 'PUT' })
    .then(function () {
      var badge = document.querySelector('.business-item[data-business-id="' + businessId + '"] .unread-badge');
      if (badge) badge.remove();
    })
    .catch(console.error);
}

function startMessagePolling() {
  pollingInterval = setInterval(function () {
    fetch(`/client/businesses/${businessId}/messages`)
      .then(r => r.text())
      .then(html => {
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        const newMsgs = doc.getElementById('messages-container');
        const curMsgs = document.getElementById('messages-container');
        if (newMsgs && curMsgs && newMsgs.innerHTML !== curMsgs.innerHTML) {
          curMsgs.innerHTML = newMsgs.innerHTML;
          curMsgs.scrollTop = curMsgs.scrollHeight;
          markAsRead();
        }
      })
      .catch(console.error);
  }, 5000);
}

function addOrderMessageToChat(order) {
  const container = document.getElementById('messages-container');
  if (!container) return;
  const div = document.createElement('div');
  div.className = 'flex justify-end';
  div.innerHTML = `<div class="max-w-xs lg:max-w-md w-full">
    <div class="bg-blue-50 border border-blue-200 rounded-lg px-4 py-3" data-message-id="${order.id}" data-order-id="${order.id}">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center space-x-2">
          <i class="fas fa-shopping-cart text-blue-600"></i>
          <span class="font-semibold text-blue-800 text-sm">[${order.id}]</span>
          <span class="text-gray-700 text-sm">${order.product_name || 'Product'}</span>
        </div>
        <button onclick="openEditOrderModal(${order.id})" class="text-blue-600 hover:text-blue-800 text-xs" title="Edit Order">
          <i class="fas fa-edit"></i>
        </button>
      </div>
      <div class="order-details text-sm text-gray-700">
        <p class="text-sm">Order #${order.order_number} - ${order.quantity || 1}x - $${parseFloat(order.total_amount).toFixed(2)}</p>
        <p class="hidden order-notes-data">${order.notes || ''}</p>
      </div>
      <div class="flex items-center justify-between mt-2">
        <p class="text-xs text-gray-500">${new Date().toLocaleTimeString('en-US', {hour:'numeric', minute:'2-digit'})}</p>
        <span class="text-xs bg-yellow-100 text-yellow-800 px-2 py-1 rounded">Pending</span>
      </div>
    </div>
  </div>`;
  container.appendChild(div);
  container.scrollTop = container.scrollHeight;
}

function addBookingMessageToChat(booking) {
  const container = document.getElementById('messages-container');
  if (!container) return;
  const bookingDate = new Date(booking.scheduled_date);
  const bookingNumber = booking.booking_number || booking.id;
  const serviceName = booking.service_name || '';
  const duration = booking.duration || '';
  const totalAmount = booking.total_amount || '';
  const notes = booking.notes || '';
  const status = booking.status || 'pending';
  const dateStr = bookingDate.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
  const timeStr = bookingDate.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
  const statusClass = status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
    status === 'confirmed' ? 'bg-green-100 text-green-800' :
    status === 'completed' ? 'bg-blue-100 text-blue-800' :
    status === 'cancelled' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800';
  const borderClass = status === 'pending' ? 'border-yellow-200 bg-yellow-50' :
    status === 'confirmed' ? 'border-green-200 bg-green-50' :
    status === 'completed' ? 'border-blue-200 bg-blue-50' :
    status === 'cancelled' ? 'border-red-200 bg-red-50' : 'border-gray-200 bg-gray-50';
  const iconClass = status === 'pending' ? 'text-yellow-600' :
    status === 'confirmed' ? 'text-green-600' :
    status === 'completed' ? 'text-blue-600' :
    status === 'cancelled' ? 'text-red-600' : 'text-gray-600';

  let extraHtml = '';
  if (status === 'pending') {
    extraHtml = '<div class="mt-2 pt-2 border-t border-gray-200 border-opacity-50"><p class="text-xs text-center text-yellow-600 font-medium"><i class="fas fa-clock mr-1"></i>Awaiting business confirmation</p></div>';
  } else if (status === 'confirmed') {
    extraHtml = '<div class="mt-2 pt-2 border-t border-gray-200 border-opacity-50"><p class="text-xs text-center text-green-600 font-medium"><i class="fas fa-check-circle mr-1"></i>Your booking is confirmed</p></div>';
  } else if (status === 'completed') {
    extraHtml = '<div class="mt-2 pt-2 border-t border-gray-200 border-opacity-50"><p class="text-xs text-center text-blue-600 font-medium"><i class="fas fa-check-double mr-1"></i>Service completed</p></div>';
  } else if (status === 'cancelled') {
    extraHtml = '<div class="mt-2 pt-2 border-t border-gray-200 border-opacity-50"><p class="text-xs text-center text-red-600 font-medium"><i class="fas fa-ban mr-1"></i>This booking was cancelled</p></div>';
  }

  container.insertAdjacentHTML('beforeend', `
    <div class="flex justify-end">
      <div class="max-w-xs lg:max-w-md w-full">
        <div class="${borderClass} border rounded-lg px-4 py-3" data-message-id="${booking.id}" data-booking-id="${booking.id}">
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center space-x-2 min-w-0">
              <i class="fas fa-calendar-check ${iconClass}"></i>
              <span class="font-semibold text-sm ${iconClass}">#${bookingNumber}</span>
              <span class="text-gray-700 text-sm truncate">${serviceName}</span>
            </div>
            <div class="flex items-center space-x-1 flex-shrink-0 ml-2">
              ${status === 'pending' ? '<button onclick="cancelBooking(' + booking.id + ')" class="text-red-500 hover:text-red-700 text-xs" title="Cancel Booking"><i class="fas fa-times"></i></button>' : ''}
              <button onclick="openEditBookingModal(${booking.id})" class="${iconClass} hover:opacity-80 text-xs" title="Edit Booking">
                <i class="fas fa-edit"></i>
              </button>
            </div>
          </div>
          <div class="booking-details text-sm text-gray-700 space-y-1">
            <p class="flex items-center space-x-1">
              <i class="fas fa-clock text-xs text-gray-400"></i>
              <span>${dateStr} at ${timeStr}</span>
            </p>
            ${duration ? '<p class="flex items-center space-x-1"><i class="fas fa-hourglass-half text-xs text-gray-400"></i><span>' + duration + ' min</span></p>' : ''}
            ${totalAmount ? '<p class="flex items-center space-x-1"><i class="fas fa-tag text-xs text-gray-400"></i><span>$' + parseFloat(totalAmount).toFixed(2) + '</span></p>' : ''}
            ${notes ? '<p class="text-xs text-gray-500 italic mt-1 border-t border-gray-200 pt-1">' + notes + '</p>' : ''}
            <p class="hidden booking-notes-data">${notes}</p>
          </div>
          <div class="flex items-center justify-between mt-3 pt-2 border-t border-gray-200 border-opacity-50">
            <p class="text-xs text-gray-400">Just now</p>
            <span class="text-xs font-medium ${statusClass} px-2 py-0.5 rounded-full booking-status">${status}</span>
          </div>
          ${extraHtml}
        </div>
      </div>
    </div>`);
  container.scrollTop = container.scrollHeight;
}

function openEditOrderModal(orderId) {
  const el = document.querySelector(`[data-order-id="${orderId}"]`) || document.querySelector(`[data-message-id="${orderId}"]`);
  if (!el) return showNotification('Order not found', 'error');
  const orderIdFromEl = el.getAttribute('data-order-id');
  if (!orderIdFromEl) return showNotification('Order ID not found', 'error');
  const notes = el.querySelector('.order-notes-data')?.textContent.trim() || '';
  const qtyEl = el.querySelector('.order-quantity-data');
  const statusEl = el.querySelector('.order-status-badge');
  const quantity = qtyEl ? parseInt(qtyEl.textContent.trim()) || 1 : 1;
  const orderNumber = el.querySelector('h4')?.textContent?.trim() || `#${orderIdFromEl}`;
  const orderStatus = statusEl ? statusEl.textContent.trim() : 'pending';

  document.getElementById('editOrderId').value = orderIdFromEl;
  document.getElementById('editOrderNotes').value = notes;
  document.getElementById('editOrderQuantity').value = quantity;

  const numberEl = document.getElementById('editOrderNumber');
  if (numberEl) numberEl.textContent = orderNumber;

  const statusBadge = document.getElementById('editOrderStatusBadge');
  if (statusBadge) {
    statusBadge.textContent = orderStatus;
    const statusColors = {
      draft: 'bg-gray-200 text-gray-600',
      pending: 'bg-blue-100 text-blue-700',
      client_confirmed: 'bg-yellow-100 text-yellow-700',
      confirmed: 'bg-green-100 text-green-700',
      fulfilled: 'bg-teal-100 text-teal-700',
      cancelled: 'bg-red-100 text-red-700'
    };
    statusBadge.className = 'text-xs font-medium px-2 py-0.5 rounded-full ' + (statusColors[orderStatus] || 'bg-gray-100 text-gray-600');
  }

  document.getElementById('editOrderModal').classList.remove('hidden');
}

function hideEditOrderModal() {
  document.getElementById('editOrderModal').classList.add('hidden');
}

function submitEditOrder() {
  const orderId = document.getElementById('editOrderId').value;
  const fd = new URLSearchParams();
  fd.append('notes', document.getElementById('editOrderNotes').value);
  fd.append('quantity', document.getElementById('editOrderQuantity').value);
  fetch(`/client/orders/${orderId}/update`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: fd
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Order updated successfully!', 'success');
        hideEditOrderModal();
      } else {
        showNotification(data.error || 'Failed to update order', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to update order', 'error'); });
}

function openEditBookingModal(messageId) {
  const el = document.querySelector(`[data-message-id="${messageId}"]`);
  if (!el) return;
  const bookingId = el.getAttribute('data-booking-id');
  if (!bookingId) return showNotification('Booking ID not found', 'error');
  const notes = el.querySelector('.booking-notes-data')?.textContent.trim() || '';
  const scheduledDate = el.querySelector('.booking-date-data')?.textContent.trim() || '';
  const numberEl = el.querySelector('.booking-number-data');
  const statusEl = el.querySelector('.booking-status');
  const bookingNumber = numberEl ? numberEl.textContent.trim() : `#${bookingId}`;
  const bookingStatus = statusEl ? statusEl.textContent.trim() : 'pending';

  document.getElementById('editBookingId').value = bookingId;
  document.getElementById('editBookingNotes').value = notes;
  if (scheduledDate) {
    try {
      const d = new Date(scheduledDate);
      document.getElementById('editBookingDate').value = d.toISOString().split('T')[0];
      document.getElementById('editBookingTime').value = d.toISOString().split('T')[1].substring(0, 5);
    } catch (e) { console.error('Error parsing date:', e); }
  }

  const numberElModal = document.getElementById('editBookingNumber');
  if (numberElModal) numberElModal.textContent = bookingNumber;

  const statusBadge = document.getElementById('editBookingStatusBadge');
  if (statusBadge) {
    statusBadge.textContent = bookingStatus;
    const statusColors = {
      pending: 'bg-yellow-100 text-yellow-800',
      confirmed: 'bg-green-100 text-green-800',
      completed: 'bg-blue-100 text-blue-800',
      cancelled: 'bg-red-100 text-red-800'
    };
    statusBadge.className = 'text-xs font-medium px-2 py-0.5 rounded-full ' + (statusColors[bookingStatus] || 'bg-gray-100 text-gray-800');
  }

  document.getElementById('editBookingModal').classList.remove('hidden');
}

function hideEditBookingModal() {
  document.getElementById('editBookingModal').classList.add('hidden');
}

function submitEditBooking() {
  const bookingId = document.getElementById('editBookingId').value;
  const fd = new URLSearchParams();
  fd.append('notes', document.getElementById('editBookingNotes').value);
  const dateVal = document.getElementById('editBookingDate').value;
  const timeVal = document.getElementById('editBookingTime').value;
  if (dateVal && timeVal) fd.append('scheduled_date', `${dateVal}T${timeVal}:00Z`);
  fetch(`/client/bookings/${bookingId}/update`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: fd
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Booking updated successfully!', 'success');
        hideEditBookingModal();
      } else {
        showNotification(data.error || 'Failed to update booking', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to update booking', 'error'); });
}

function clientConfirmOrder(orderId) {
  if (!confirm('Confirm this order?')) return;
  fetch(`/client/orders/${orderId}/confirm`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ items: [] })
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification(data.message || 'Order confirmed!', 'success');
        // Trigger polling refresh
        setTimeout(() => {
          fetch(`/client/businesses/${businessId}/messages`)
            .then(r => r.text())
            .then(html => {
              const parser = new DOMParser();
              const doc = parser.parseFromString(html, 'text/html');
              const newMsgs = doc.getElementById('messages-container');
              const curMsgs = document.getElementById('messages-container');
              if (newMsgs && curMsgs) {
                curMsgs.innerHTML = newMsgs.innerHTML;
              }
            });
        }, 500);
      } else {
        showNotification(data.error || 'Failed to confirm order', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to confirm order', 'error'); });
}

function clientRequestChanges(orderId) {
  const message = prompt('Describe the changes you need:');
  if (!message) return;
  // For now, just send a text message requesting changes
  const formData = new FormData();
  formData.append('content', 'Request changes for order #' + orderId + ': ' + message);
  formData.append('sender', 'client');
  htmx.ajax('POST', `/client/businesses/${businessId}/messages`, {
    values: formData,
    target: '#messages-container',
    swap: 'beforeend'
  });
}

function clientOrderItemIncrement(orderId, productId, btn) {
  const qtySpan = btn.parentElement.querySelector('.qty-value');
  const current = parseInt(qtySpan.textContent);
  qtySpan.textContent = current + 1;
  updateClientOrderTotal(orderId);
}

function clientOrderItemDecrement(orderId, productId, btn) {
  const qtySpan = btn.parentElement.querySelector('.qty-value');
  const current = parseInt(qtySpan.textContent);
  if (current > 1) {
    qtySpan.textContent = current - 1;
  }
  updateClientOrderTotal(orderId);
}

function updateClientOrderTotal(orderId) {
  const card = document.querySelector(`[data-order-id="${orderId}"]`);
  if (!card) return;
  let total = 0;
  card.querySelectorAll('[data-item-product-id]').forEach(item => {
    const qty = parseInt(item.querySelector('.qty-value').textContent);
    const priceEl = item.closest('.flex.items-center.justify-between').querySelector('.text-sm.font-bold');
    const priceText = priceEl ? priceEl.textContent.replace('$', '') : '0';
    total += qty * parseFloat(priceText);
  });
  const totalEl = card.querySelector('.text-lg.font-bold.text-gray-900');
  if (totalEl) totalEl.textContent = '$' + total.toFixed(2);
}

function cancelOrder(orderId) {
  if (!confirm('Are you sure you want to cancel this order?')) return;
  fetch(`/client/orders/${orderId}/cancel`, {
    method: 'POST',
    headers: { 'Authorization': 'Bearer ' + getCookie('client_token') }
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Order cancelled successfully', 'success');
      } else {
        showNotification(data.error || 'Failed to cancel order', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to cancel order', 'error'); });
}

function cancelBooking(bookingId) {
  if (!confirm('Are you sure you want to cancel this booking?')) return;
  fetch(`/client/bookings/${bookingId}/cancel`, {
    method: 'POST',
    headers: { 'Authorization': 'Bearer ' + getCookie('client_token') }
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Booking cancelled successfully', 'success');
      } else {
        showNotification(data.error || 'Failed to cancel booking', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to cancel booking', 'error'); });
}
