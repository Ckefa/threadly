let pollingInterval = null;

scrollToBottom();
markAsRead();
startMessagePolling();

function scrollToBottom() {
  var container = document.getElementById('messages-container');
  if (container) requestAnimationFrame(function () {
    container.scrollTop = container.scrollHeight;
  });
}

function markAsRead() {
  fetch(`/business/clients/${clientId}/read`, { method: 'PUT' })
    .then(function () {
      var badge = document.querySelector('.client-item[data-client-id="' + clientId + '"] .unread-badge');
      if (badge) badge.remove();
    })
    .catch(console.error);
}

function startMessagePolling() {
  pollingInterval = setInterval(function () {
    fetchMessages();
  }, 5000);
}

function fetchMessages() {
  fetch(`clients/${clientId}/messages`)
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

window.addEventListener('beforeunload', function () {
  if (pollingInterval) {
    clearInterval(pollingInterval);
  }
});

document.addEventListener('click', function (e) {
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

  fetch('/clients/' + clientId + '/status', {
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
  fetch('/clients/' + clientId + '/conversation-id')
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

  htmx.ajax('POST', '/messages/' + messageId + '/actions', {
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

  fetch('/conversations/' + conversationId + '/status', {
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

  fetch('/conversations/' + conversationId + '/status', {
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
  fetch(`/messages/${messageId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({content: newContent})
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
  const notes = notesEl ? notesEl.textContent.trim() : '';

  document.getElementById('editOrderId').value = orderIdFromEl;
  document.getElementById('editOrderNotes').value = notes;
  document.getElementById('editOrderQuantity').value = 1;
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

  const notes = notesEl ? notesEl.textContent.trim() : '';
  const date = dateEl ? dateEl.textContent.trim() : '';
  const time = timeEl ? timeEl.textContent.trim() : '';

  document.getElementById('editBookingId').value = bookingIdFromEl;
  document.getElementById('editBookingNotes').value = notes;
  document.getElementById('editBookingDate').value = date;
  document.getElementById('editBookingTime').value = time;
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

  const bookingDetails = bookingCard.querySelector('.booking-details p');
  if (bookingDetails && updatedBooking.scheduled_date) {
    const date = new Date(updatedBooking.scheduled_date);
    bookingDetails.textContent = date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hour12: true
    });
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
    statusBadge.textContent = updatedBooking.status;
  }
}

function submitEditOrder() {
  const orderId = document.getElementById('editOrderId').value;
  const notes = document.getElementById('editOrderNotes').value;
  const quantity = document.getElementById('editOrderQuantity').value;

  const formData = new URLSearchParams();
  formData.append('notes', notes);
  formData.append('quantity', quantity);

  fetch(`/client/orders/${orderId}/update`, {
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
      } else {
        showNotification(data.error || 'Failed to update order', 'error');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      showNotification('Failed to update order', 'error');
    });
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

  fetch(`/bookings/${bookingId}`, {
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
