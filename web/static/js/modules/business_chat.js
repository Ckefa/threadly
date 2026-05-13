let pollingInterval = null;

scrollToBottom();
markAsRead();
startMessagePolling();

function scrollToBottom() {
  var container = document.getElementById('messages-container');
  if (container) requestAnimationFrame(function() {
    container.scrollTop = container.scrollHeight;
  });
}

function markAsRead() {
  fetch(`/business/clients/${clientId}/read`, { method: 'PUT' })
    .then(function() {
      var badge = document.querySelector('.client-item[data-client-id="' + clientId + '"] .unread-badge');
      if (badge) badge.remove();
    })
    .catch(console.error);
}

function startMessagePolling() {
  pollingInterval = setInterval(function() {
    fetchMessages();
  }, 5000);
}

function fetchMessages() {
  fetch(`/business/clients/${clientId}/messages`)
    .then(response => response.text())
    .then(html => {
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');
      const newMessages = doc.getElementById('messages-container');
      const currentMessages = document.getElementById('messages-container');

      if (newMessages && currentMessages && newMessages.innerHTML !== currentMessages.innerHTML) {
        currentMessages.innerHTML = newMessages.innerHTML;
        currentMessages.scrollTop = currentMessages.scrollHeight;
        markAsRead();
      }
    })
    .catch(error => {
      console.error('Error fetching messages:', error);
    });
}

window.addEventListener('beforeunload', function() {
  if (pollingInterval) {
    clearInterval(pollingInterval);
  }
});

document.addEventListener('click', function(e) {
  const progressBtn = e.target.closest('.view-chat-progress-btn');
  if (progressBtn) {
    const clientId = progressBtn.getAttribute('data-client-id');
    showConversationProgress(clientId);
  }

  const statusDropdown = e.target.closest('.conversation-status-dropdown');
  if (statusDropdown && statusDropdown.value) {
    const conversationId = statusDropdown.getAttribute('data-conversation-id');
    updateConversationStatus(conversationId, statusDropdown.value);
  }

  const saveProgressBtn = e.target.closest('.save-chat-progress-btn');
  if (saveProgressBtn) {
    const conversationId = saveProgressBtn.getAttribute('data-conversation-id');
    const dropdown = document.querySelector('.conversation-progress-dropdown[data-conversation-id="' + conversationId + '"]');
    if (dropdown && dropdown.value) {
      saveChatProgress(conversationId, dropdown.value);
    }
  }
});

function updateClientStatus(clientId, status) {
  console.log('Updating client status:', clientId, status);
  const formData = new FormData();
  formData.append('status', status);

  fetch('/business/clients/' + clientId + '/status', {
    method: 'PUT',
    body: formData
  })
    .then(response => {
      console.log('Customer status response status:', response.status);
      if (!response.ok) {
        throw new Error('HTTP error! status: ' + response.status);
      }
      return response.json();
    })
    .then(data => {
      console.log('Customer status response data:', data);
      if (data.client) {
        showNotification('Customer status updated to ' + status, 'success');
        const dropdown = document.querySelector('select[data-client-id="' + clientId + '"]');
        if (dropdown) {
          dropdown.value = status;
          console.log('Updated customer dropdown to:', status);
        }
      } else {
        console.error('No client data in response');
        showNotification('Failed to update client status', 'error');
      }
    })
    .catch(error => {
      console.error('Error updating client status:', error);
      showNotification('Failed to update client status: ' + error.message, 'error');
    });
}

function showEnhancedActionModal(messageId) {
  htmx.ajax('GET', '/actions/modal/' + messageId, {
    target: 'body',
    swap: 'beforeend'
  });
}

function showConversationProgress(clientId) {
  fetch('/business/clients/' + clientId + '/conversation-id')
    .then(response => response.json())
    .then(data => {
      if (data.conversation_id) {
        htmx.ajax('GET', '/conversations/' + data.conversation_id + '/progress', {
          target: '#progress-modal',
          swap: 'innerHTML'
        });
        showProgressModal();
      }
    })
    .catch(console.error);
}

function showProgressModal() {
  if (!document.getElementById('progress-modal')) {
    const modal = document.createElement('div');
    modal.id = 'progress-modal';
    modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
    document.body.appendChild(modal);
  }
}

function createQuickAction(messageId, actionType) {
  const formData = new FormData();
  formData.append('message_id', messageId);
  formData.append('type', actionType);
  formData.append('title', actionType.charAt(0).toUpperCase() + actionType.slice(1) + ' from message');

  htmx.ajax('POST', 'business/messages/' + messageId + '/actions', {
    values: formData,
    target: '#action-result',
    swap: 'innerHTML'
  });
}

function updateConversationStatus(conversationId, stage) {
  console.log('Updating conversation status:', conversationId, stage);
  const formData = new FormData();
  formData.append('stage', stage);
  formData.append('reason', 'Manual status update by business');

  fetch('/business/conversations/' + conversationId + '/status', {
    method: 'PUT',
    body: formData
  })
    .then(response => {
      console.log('Response status:', response.status);
      if (!response.ok) {
        throw new Error('HTTP error! status: ' + response.status);
      }
      return response.json();
    })
    .then(data => {
      console.log('Response data:', data);
      if (data.progress) {
        showNotification('Conversation stage updated to ' + stage, 'success');
        const dropdown = document.querySelector('.conversation-status-dropdown');
        if (dropdown) {
          dropdown.value = stage;
          console.log('Updated dropdown to:', stage);
        }
      } else {
        console.error('No progress data in response');
        showNotification('Failed to update conversation stage', 'error');
      }
    })
    .catch(error => {
      console.error('Error updating conversation status:', error);
      showNotification('Failed to update conversation status: ' + error.message, 'error');
    });
}

function saveChatProgress(conversationId, stage) {
  console.log('Saving chat progress:', conversationId, stage);
  const formData = new FormData();
  formData.append('stage', stage);
  formData.append('reason', 'Manual progress update from chat');

  fetch('/business/conversations/' + conversationId + '/status', {
    method: 'PUT',
    body: formData
  })
    .then(response => {
      console.log('Chat progress response status:', response.status);
      if (!response.ok) {
        throw new Error('HTTP error! status: ' + response.status);
      }
      return response.json();
    })
    .then(data => {
      console.log('Chat progress response data:', data);
      if (data.progress) {
        showNotification('Conversation progress saved: ' + stage, 'success');
        const progressScoreElement = document.querySelector('.text-xs.text-gray-500');
        if (progressScoreElement && data.progress.progress_score) {
          progressScoreElement.textContent = 'Progress Score: ' + data.progress.progress_score + '%';
        }
      } else {
        console.error('No progress data in response');
        showNotification('Failed to save conversation progress', 'error');
      }
    })
    .catch(error => {
      console.error('Error saving chat progress:', error);
      showNotification('Failed to save conversation progress: ' + error.message, 'error');
    });
}

function editOrderMessage(messageId) {
  const messageEl = document.querySelector(`[data-message-id="${messageId}"]`);
  if (!messageEl) return;

  const currentContent = messageEl.querySelector('.order-details p')?.textContent || '';

  const parts = currentContent.split('|').map(p => p.trim());
  let notes = '';
  const notesPart = parts.find(p => p.startsWith('Notes:'));
  if (notesPart) {
    notes = notesPart.replace('Notes:', '').trim();
  }

  const newNotes = prompt('Update order notes:', notes);
  if (newNotes === null) return;

  let newContent = parts.filter(p => !p.startsWith('Notes:')).join(' | ');
  if (newNotes.trim()) {
    newContent += ` | Notes: ${newNotes.trim()}`;
  }

  updateMessageContent(messageId, newContent);
}

function editBookingMessage(messageId) {
  const messageEl = document.querySelector(`[data-message-id="${messageId}"]`);
  if (!messageEl) return;

  const currentContent = messageEl.querySelector('.booking-details p')?.textContent || '';

  const parts = currentContent.split('|').map(p => p.trim());
  let notes = '';
  const notesPart = parts.find(p => p.startsWith('Notes:'));
  if (notesPart) {
    notes = notesPart.replace('Notes:', '').trim();
  }

  const newNotes = prompt('Update booking notes:', notes);
  if (newNotes === null) return;

  let newContent = parts.filter(p => !p.startsWith('Notes:')).join(' | ');
  if (newNotes.trim()) {
    newContent += ` | Notes: ${newNotes.trim()}`;
  }

  updateMessageContent(messageId, newContent);
}

function updateMessageContent(messageId, newContent) {
  fetch(`/business/messages/${messageId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ content: newContent })
  })
    .then(response => response.json())
    .then(data => {
      if (data.success) {
        showNotification('Updated successfully!', 'success');
        fetchMessages();
      } else {
        showNotification(data.error || 'Failed to update', 'error');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      showNotification('Failed to update', 'error');
    });
}

function showNotification(message, type) {
  const notification = document.createElement('div');
  notification.className = `fixed top-4 right-4 px-4 py-2 rounded-lg text-white z-50 ${type === 'success' ? 'bg-green-500' : 'bg-red-500'}`;
  notification.textContent = message;

  document.body.appendChild(notification);

  setTimeout(() => {
    notification.remove();
  }, 3000);
}

function openEditOrderModal(orderId) {
  const messageEl = document.querySelector(`[data-order-id="${orderId}"]`);
  if (!messageEl) return;

  const orderIdFromEl = messageEl.getAttribute('data-order-id');
  if (!orderIdFromEl) {
    showNotification('Order ID not found', 'error');
    return;
  }

  const notesEl = messageEl.querySelector('.order-notes-data');
  const qtyEl = messageEl.querySelector('.order-quantity-data');
  const statusEl = messageEl.querySelector('.order-status-badge');
  const notes = notesEl ? notesEl.textContent.trim() : '';
  const quantity = qtyEl ? parseInt(qtyEl.textContent.trim()) || 1 : 1;
  const orderNumber = messageEl.querySelector('h4')?.textContent?.trim() || `#${orderIdFromEl}`;
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

function openEditBookingModal(bookingId) {
  const messageEl = document.querySelector(`[data-booking-id="${bookingId}"]`);
  if (!messageEl) return;

  const bookingIdFromEl = messageEl.getAttribute('data-booking-id');
  if (!bookingIdFromEl) {
    showNotification('Booking ID not found', 'error');
    return;
  }

  const notesEl = messageEl.querySelector('.booking-notes-data');
  const dateEl = messageEl.querySelector('.booking-date-data');
  const timeEl = messageEl.querySelector('.booking-time-data');
  const numberEl = messageEl.querySelector('.booking-number-data');
  const statusEl = messageEl.querySelector('.booking-status');

  const notes = notesEl ? notesEl.textContent.trim() : '';
  const date = dateEl ? dateEl.textContent.trim() : '';
  const time = timeEl ? timeEl.textContent.trim() : '';
  const bookingNumber = numberEl ? numberEl.textContent.trim() : `#${bookingIdFromEl}`;
  const bookingStatus = statusEl ? statusEl.textContent.trim() : 'pending';

  document.getElementById('editBookingId').value = bookingIdFromEl;
  document.getElementById('editBookingNotes').value = notes;
  document.getElementById('editBookingDate').value = date;
  document.getElementById('editBookingTime').value = time;

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

function updateOrderCard(orderId, updatedOrder) {
  const orderCard = document.querySelector(`[data-order-id="${orderId}"]`);
  if (!orderCard) return;

  const orderDetails = orderCard.querySelector('.order-details p');
  if (orderDetails) {
    const quantity = updatedOrder.quantity || 1;
    const totalAmount = parseFloat(updatedOrder.total_amount).toFixed(2);
    orderDetails.textContent = `Order #${updatedOrder.order_number} - ${quantity}x - $${totalAmount}`;
  }

  const notesData = orderCard.querySelector('.order-notes-data');
  if (notesData) {
    notesData.textContent = updatedOrder.notes || '';
  }

  const statusBadge = orderCard.querySelector('.order-status');
  if (statusBadge && updatedOrder.status) {
    statusBadge.textContent = updatedOrder.status;
  }
}

function updateBookingCard(bookingId, updatedBooking) {
  const bookingCard = document.querySelector(`[data-booking-id="${bookingId}"]`);
  if (!bookingCard) return;

  const bookingDetails = bookingCard.querySelector('.booking-details');
  if (bookingDetails && updatedBooking.scheduled_date) {
    const date = new Date(updatedBooking.scheduled_date);
    const dateStr = date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
    const timeStr = date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
    const timeSpan = bookingDetails.querySelector('.flex.items-center.space-x-1 span');
    if (timeSpan) {
      timeSpan.textContent = `${dateStr} at ${timeStr}`;
    }
  }

  const notesData = bookingCard.querySelector('.booking-notes-data');
  if (notesData) {
    notesData.textContent = updatedBooking.notes || '';
  }

  const dateData = bookingCard.querySelector('.booking-date-data');
  const timeData = bookingCard.querySelector('.booking-time-data');
  if (dateData && updatedBooking.scheduled_date) {
    const date = new Date(updatedBooking.scheduled_date);
    dateData.textContent = date.toISOString().split('T')[0];
    timeData.textContent = date.toISOString().split('T')[1].substring(0, 5);
  }

  const statusBadge = bookingCard.querySelector('.booking-status');
  if (statusBadge && updatedBooking.status) {
    const status = updatedBooking.status;
    statusBadge.textContent = status;
    const statusColors = {
      pending: 'bg-yellow-100 text-yellow-800',
      confirmed: 'bg-green-100 text-green-800',
      completed: 'bg-blue-100 text-blue-800',
      cancelled: 'bg-red-100 text-red-800'
    };
    statusBadge.className = 'text-xs font-medium px-2 py-0.5 rounded-full booking-status ' + (statusColors[status] || 'bg-gray-100 text-gray-800');
  }

  const actionBtns = bookingCard.querySelector('.flex.items-center.space-x-2.mt-2');
  if (actionBtns) {
    const status = updatedBooking.status;
    if (status === 'pending') {
      actionBtns.innerHTML = `
        <button onclick="updateBookingStatusFromCard(${bookingId}, 'confirmed')" class="flex-1 text-xs bg-green-100 text-green-700 hover:bg-green-200 px-2 py-1 rounded font-medium"><i class="fas fa-check mr-1"></i>Confirm</button>
        <button onclick="updateBookingStatusFromCard(${bookingId}, 'cancelled')" class="flex-1 text-xs bg-red-100 text-red-700 hover:bg-red-200 px-2 py-1 rounded font-medium"><i class="fas fa-times mr-1"></i>Cancel</button>`;
    } else if (status === 'confirmed') {
      actionBtns.innerHTML = `
        <button onclick="updateBookingStatusFromCard(${bookingId}, 'completed')" class="flex-1 text-xs bg-blue-100 text-blue-700 hover:bg-blue-200 px-2 py-1 rounded font-medium"><i class="fas fa-check-double mr-1"></i>Complete</button>
        <button onclick="updateBookingStatusFromCard(${bookingId}, 'cancelled')" class="flex-1 text-xs bg-red-100 text-red-700 hover:bg-red-200 px-2 py-1 rounded font-medium"><i class="fas fa-times mr-1"></i>Cancel</button>`;
    } else {
      actionBtns.remove();
    }
  }
}

function submitEditOrder() {
  const orderId = document.getElementById('editOrderId').value;
  const notes = document.getElementById('editOrderNotes').value;
  const quantity = document.getElementById('editOrderQuantity').value;

  const formData = new URLSearchParams();
  formData.append('notes', notes);
  formData.append('quantity', quantity);

  fetch(`/business/client/orders/${orderId}/update`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: formData
  })
    .then(response => response.json())
    .then(data => {
      if (data.success) {
        showNotification('Order updated successfully!', 'success');
        hideEditOrderModal();
        updateOrderCard(orderId, data.order);
        fetchMessages();
      } else {
        showNotification(data.error || 'Failed to update order', 'error');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      showNotification('Failed to update order', 'error');
    });
}

// ========== Order Lifecycle Functions ==========

function sendOrderToClient(orderId) {
  if (!confirm('Send this order to the client?')) return;
  fetch(`/business/orders/${orderId}/send`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Order sent to client!', 'success');
        fetchMessages();
      } else {
        showNotification(data.error || 'Failed to send order', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to send order', 'error'); });
}

function confirmOrderBusiness(orderId) {
  if (!confirm('Confirm and process this order?')) return;
  fetch(`/business/orders/${orderId}/confirm`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Order confirmed!', 'success');
        fetchMessages();
      } else {
        showNotification(data.error || 'Failed to confirm order', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to confirm order', 'error'); });
}

function rejectOrder(orderId) {
  if (!confirm('Reject this order?')) return;
  fetch(`/business/orders/${orderId}/reject`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Order rejected', 'info');
        fetchMessages();
      } else {
        showNotification(data.error || 'Failed to reject order', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to reject order', 'error'); });
}

function cancelDraftOrder(orderId) {
  if (!confirm('Discard this draft order?')) return;
  fetch(`/business/orders/${orderId}/reject`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification('Draft discarded', 'info');
        fetchMessages();
      } else {
        showNotification(data.error || 'Failed to discard draft', 'error');
      }
    })
    .catch(e => { console.error(e); showNotification('Failed to discard draft', 'error'); });
}

function orderItemIncrement(orderId, productId, btn) {
  fetch(`/business/orders/${orderId}/items/${productId}/increment`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.success) fetchMessages();
    })
    .catch(console.error);
}

function orderItemDecrement(orderId, productId, btn) {
  fetch(`/business/orders/${orderId}/items/${productId}/decrement`, { method: 'POST' })
    .then(r => r.json())
    .then(data => {
      if (data.success) fetchMessages();
    })
    .catch(console.error);
}

function updateBookingStatusFromCard(bookingId, newStatus) {
  const action = newStatus === 'confirmed' ? 'confirm' : newStatus === 'completed' ? 'complete' : 'cancel';
  if (!confirm(`Are you sure you want to ${action} this booking?`)) return;

  fetch(`/business/bookings/${bookingId}/status`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status: newStatus })
  })
    .then(r => r.json())
    .then(data => {
      if (data.success) {
        showNotification(`Booking ${action}ed successfully!`, 'success');
        fetchMessages();
      } else {
        showNotification(data.error || `Failed to ${action} booking`, 'error');
      }
    })
    .catch(e => { console.error(e); showNotification(`Failed to ${action} booking`, 'error'); });
}

function submitEditBooking() {
  const bookingId = document.getElementById('editBookingId').value;
  const notes = document.getElementById('editBookingNotes').value;
  const date = document.getElementById('editBookingDate').value;
  const time = document.getElementById('editBookingTime').value;

  const requestData = {
    notes: notes,
    booking_date: `${date}T${time}:00Z`
  };

  fetch(`/business/bookings/${bookingId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(requestData)
  })
    .then(response => response.json())
    .then(data => {
      if (data.success) {
        showNotification('Booking updated successfully!', 'success');
        hideEditBookingModal();
        updateBookingCard(bookingId, data.booking);
      } else {
        showNotification(data.error || 'Failed to update booking', 'error');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      showNotification('Failed to update booking', 'error');
    });
}

// ========== Message Search ==========

function toggleMessageSearch() {
  var bar = document.getElementById('messageSearchBar');
  if (bar) bar.classList.toggle('hidden');
  if (!bar.classList.contains('hidden')) {
    setTimeout(function() {
      document.getElementById('messageSearchInput')?.focus();
    }, 100);
  } else {
    clearMessageSearch();
  }
}

function filterMessages(query) {
  var q = query.toLowerCase().trim();
  var container = document.getElementById('messages-container');
  var messages = container.querySelectorAll(':scope > div');
  var count = 0;
  messages.forEach(function(el) {
    var text = el.getAttribute('data-message-text') || el.textContent.toLowerCase();
    if (!q || text.toLowerCase().includes(q)) {
      el.style.display = '';
      count++;
    } else {
      el.style.display = 'none';
    }
  });
  var countEl = document.getElementById('searchResultCount');
  if (countEl) countEl.textContent = count;
}

function clearMessageSearch() {
  var input = document.getElementById('messageSearchInput');
  if (input) input.value = '';
  var container = document.getElementById('messages-container');
  if (container) {
    container.querySelectorAll(':scope > div').forEach(function(el) {
      el.style.display = '';
    });
  }
  var countEl = document.getElementById('searchResultCount');
  if (countEl) countEl.textContent = '0';
}

// ========== Customer Info Panel ==========

function toggleCustomerInfo() {
  var panel = document.getElementById('customerInfoPanel');
  if (panel) panel.classList.toggle('hidden');
}

// ========== Smart Suggestions ==========

function hideSmartSuggestions() {
  var bar = document.getElementById('smartSuggestions');
  if (bar) bar.classList.add('hidden');
}

// ========== Quick Replies & Input Handling ==========

function onMessageInput(input) {
  var val = input.value;
  var suggestions = document.getElementById('smartSuggestions');
  if (suggestions) {
    if (val.length > 0) {
      suggestions.classList.remove('hidden');
    } else {
      suggestions.classList.add('hidden');
    }
  }

  // Show quick replies when typing /
  var qr = document.getElementById('quickReplies');
  if (qr) {
    if (val === '/') {
      qr.classList.remove('hidden');
    } else if (qr && !qr.classList.contains('hidden') && val.charAt(0) !== '/') {
      qr.classList.add('hidden');
    }
  }
}

function onMessageKeydown(event) {
  var qr = document.getElementById('quickReplies');
  if (event.key === 'Escape' && qr && !qr.classList.contains('hidden')) {
    qr.classList.add('hidden');
    var input = document.getElementById('messageInput');
    if (input) input.value = input.value.replace(/\/$/, '');
  }
}

function insertQuickReply(text) {
  var input = document.getElementById('messageInput');
  if (input) {
    input.value = text;
    input.focus();
  }
  var qr = document.getElementById('quickReplies');
  if (qr) qr.classList.add('hidden');
  var suggestions = document.getElementById('smartSuggestions');
  if (suggestions) suggestions.classList.remove('hidden');
}


