

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
  container.insertAdjacentHTML('beforeend', `
    <div class="flex justify-end">
      <div class="max-w-xs lg:max-w-md w-full">
        <div class="bg-purple-50 border border-purple-200 rounded-lg px-4 py-3" data-message-id="${booking.id}" data-booking-id="${booking.id}">
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center space-x-2">
              <i class="fas fa-calendar text-purple-600"></i>
              <span class="font-semibold text-purple-800 text-sm">Booking</span>
            </div>
            <button onclick="openEditBookingModal(${booking.id})" class="text-purple-600 hover:text-purple-800 text-xs" title="Edit Booking">
              <i class="fas fa-edit"></i>
            </button>
          </div>
          <div class="booking-details text-sm text-gray-700">
            <p class="text-sm">📅 BOOKING:${booking.id} | ${booking.service_name} | ${bookingDate.toLocaleDateString()} ${bookingDate.toLocaleTimeString([], {hour:'2-digit', minute:'2-digit'})} | Status: pending${booking.notes ? ' | Notes: ' + booking.notes : ''}</p>
            <p class="hidden booking-notes-data">${booking.notes || ''}</p>
          </div>
          <div class="flex items-center justify-between mt-2">
            <p class="text-xs text-gray-500">Just now</p>
            <span class="text-xs bg-yellow-100 text-yellow-800 px-2 py-1 rounded">Pending</span>
          </div>
        </div>
      </div>
    </div>`);
  container.scrollTop = container.scrollHeight;
}

function openEditOrderModal(messageId) {
  const el = document.querySelector(`[data-message-id="${messageId}"]`);
  if (!el) return;
  const orderId = el.getAttribute('data-order-id');
  if (!orderId) return showNotification('Order ID not found', 'error');
  const notes = el.querySelector('.order-notes-data')?.textContent.trim() || '';
  document.getElementById('editOrderId').value = orderId;
  document.getElementById('editOrderNotes').value = notes;
  document.getElementById('editOrderQuantity').value = '1';
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
  document.getElementById('editBookingId').value = bookingId;
  document.getElementById('editBookingNotes').value = notes;
  if (scheduledDate) {
    try {
      const d = new Date(scheduledDate);
      document.getElementById('editBookingDate').value = d.toISOString().split('T')[0];
      document.getElementById('editBookingTime').value = d.toISOString().split('T')[1].substring(0, 5);
    } catch (e) { console.error('Error parsing date:', e); }
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
