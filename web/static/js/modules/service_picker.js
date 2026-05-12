// Service Picker - shared between business and client
let servicePickerMode = 'business';
let servicePickerConversationId = null;
let servicePickerBusinessId = null;
let servicePickerClientId = null;
let servicePickerServices = [];
let servicePickerSelectedService = null;
let servicePickerCategories = [];

// ========== ENTRY POINTS ==========

function openServicePicker(clientId) {
  if (!clientId) { showNotification('No client selected', 'error'); return; }
  fetch(`/business/clients/${clientId}/conversation-id`, { credentials: 'same-origin' })
    .then(r => r.json())
    .then(data => {
      if (data.conversation_id) {
        servicePickerMode = 'business';
        servicePickerConversationId = data.conversation_id;
        servicePickerClientId = clientId;
        showServicePicker();
      } else {
        showNotification('Could not find conversation', 'error');
      }
    })
    .catch(e => {
      console.error('Failed to get conversation:', e);
      showNotification('Failed to get conversation', 'error');
    });
}

function openClientServicePicker() {
  if (!businessId) { showNotification('No business selected', 'error'); return; }
  servicePickerMode = 'client';
  servicePickerBusinessId = businessId;
  showServicePicker();
}

// ========== MODAL CONTROLS ==========

function showServicePicker() {
  const modal = document.getElementById('servicePickerModal');
  if (!modal) { showNotification('Service picker not loaded yet', 'error'); return; }
  modal.classList.remove('hidden');
  document.body.style.overflow = 'hidden';
  servicePickerSelectedService = null;

  document.getElementById('servicePickerStep1').classList.remove('hidden');
  document.getElementById('servicePickerStep2').classList.add('hidden');
  document.getElementById('servicePickerStep3').classList.add('hidden');

  const titleEl = document.getElementById('servicePickerTitle');
  if (titleEl) titleEl.textContent = servicePickerMode === 'client' ? 'Request a Service' : 'Book a Service';

  const submitBtn = document.getElementById('servicePickerSubmitBtn');
  if (submitBtn) {
    submitBtn.innerHTML = servicePickerMode === 'client'
      ? '<i class="fas fa-paper-plane mr-1"></i> Send Request'
      : '<i class="fas fa-calendar-check mr-1"></i> Create Booking';
  }

  updateServiceStepIndicators(1);
  loadPickerServices();
}

function hideServicePicker() {
  const modal = document.getElementById('servicePickerModal');
  if (modal) modal.classList.add('hidden');
  document.body.style.overflow = '';
  servicePickerSelectedService = null;
  servicePickerServices = [];
  servicePickerCategories = [];
}

function servicePickerGoToStep(step) {
  ['servicePickerStep1', 'servicePickerStep2', 'servicePickerStep3'].forEach((id, i) => {
    const el = document.getElementById(id);
    if (el) el.classList.toggle('hidden', i + 1 !== step);
  });
  updateServiceStepIndicators(step);
  if (step === 3) renderServiceConfirmation();
}

function updateServiceStepIndicators(step) {
  [
    { ind: 'svcStep1Indicator', label: 'svcStep1Label', line: 'svcStep1to2' },
    { ind: 'svcStep2Indicator', label: 'svcStep2Label', line: 'svcStep2to3' },
    { ind: 'svcStep3Indicator', label: 'svcStep3Label', line: null }
  ].forEach((item, i) => {
    const idx = i + 1;
    const indEl = document.getElementById(item.ind);
    const labelEl = document.getElementById(item.label);
    const isActive = idx <= step;
    if (indEl) {
      indEl.className = `w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold transition-all duration-300 ${isActive ? 'bg-purple-500 text-white shadow-sm' : 'bg-gray-200 text-gray-400'}`;
      if (isActive && idx < step) indEl.innerHTML = '<i class="fas fa-check text-[10px]"></i>';
      else indEl.textContent = String(idx);
    }
    if (labelEl) {
      labelEl.className = `text-xs font-medium transition-colors duration-300 ${isActive ? 'text-purple-600' : 'text-gray-400'}`;
    }
    if (item.line) {
      const lineEl = document.getElementById(item.line);
      if (lineEl) lineEl.className = `w-12 h-px mx-2 transition-colors duration-300 ${isActive ? 'bg-purple-300' : 'bg-gray-200'}`;
    }
  });
}

// ========== SERVICE LOADING ==========

async function loadPickerServices() {
  const grid = document.getElementById('servicePickerGrid');
  const countEl = document.getElementById('servicePickerCount');
  if (!grid) return;
  grid.innerHTML = '<div class="text-center py-16 text-gray-400"><i class="fas fa-spinner fa-spin text-3xl mb-3"></i><p class="text-sm">Loading services...</p></div>';

  try {
    let url;
    if (servicePickerMode === 'client') {
      url = `/client/businesses/${servicePickerBusinessId}/services`;
    } else {
      url = `/business/conversations/${servicePickerConversationId}/services`;
    }
    const resp = await fetch(url, { credentials: 'same-origin', redirect: 'error' });

    if (!resp.ok) {
      const msg = resp.status === 401 ? 'Not authenticated.' :
                  resp.status === 404 ? 'Services not found.' :
                  `Server error (${resp.status})`;
      grid.innerHTML = `<div class="text-center py-12 text-orange-400"><i class="fas fa-exclamation-triangle text-3xl mb-3"></i><p>${msg}</p></div>`;
      return;
    }

    const data = await resp.json();
    let services;
    if (Array.isArray(data)) {
      services = data;
    } else if (data.services) {
      services = data.services;
    } else {
      services = [];
    }

    servicePickerServices = services;

    servicePickerCategories = [...new Set(services.map(s => s.category || s.Category || '').filter(Boolean))];

    if (countEl) countEl.textContent = `${services.length} service${services.length !== 1 ? 's' : ''} available`;
    renderServicePicker();
  } catch (e) {
    console.error('Failed to load services:', e);
    grid.innerHTML = '<div class="text-center py-12 text-red-400"><i class="fas fa-exclamation-triangle text-3xl mb-3"></i><p>Failed to load services</p></div>';
  }
}

function renderServicePicker(category) {
  const grid = document.getElementById('servicePickerGrid');
  const catFilter = document.getElementById('servicePickerCategoryFilter');
  if (!grid) return;

  const searchInput = document.getElementById('servicePickerSearch');
  const search = searchInput ? (searchInput.value || '').toLowerCase() : '';

  const clearBtn = document.getElementById('servicePickerSearchClear');
  if (clearBtn) clearBtn.classList.toggle('hidden', !search);

  if (!servicePickerServices || servicePickerServices.length === 0) {
    grid.innerHTML = '<div class="text-center py-16 text-gray-400"><i class="fas fa-concierge-bell text-5xl mb-3"></i><p class="text-sm font-medium">No services available</p><p class="text-xs mt-1">Check back later for new services</p></div>';
    if (catFilter) catFilter.innerHTML = '';
    return;
  }

  // Render category filters
  if (catFilter) {
    let catHtml = '<button onclick="filterServiceCategory(\'\')" class="px-3 py-1.5 rounded-full text-xs font-medium transition ' +
      (!category ? 'bg-purple-600 text-white shadow-sm' : 'bg-gray-100 text-gray-600 hover:bg-gray-200') + '">All</button>';
    servicePickerCategories.forEach(cat => {
      catHtml += `<button onclick="filterServiceCategory('${cat.replace(/'/g, "\\'")}')" class="px-3 py-1.5 rounded-full text-xs font-medium transition ${category === cat ? 'bg-purple-600 text-white shadow-sm' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}">${cat}</button>`;
    });
    catFilter.innerHTML = catHtml;
  }

  const filtered = servicePickerServices.filter(s => {
    const name = (s.name || s.Name || '').toLowerCase();
    const svcCategory = (s.category || s.Category || '').toLowerCase();
    const matchesSearch = !search || name.includes(search) || svcCategory.includes(search);
    const matchesCategory = !category || svcCategory === category.toLowerCase();
    return matchesSearch && matchesCategory;
  });

  if (filtered.length === 0) {
    grid.innerHTML = '<div class="text-center py-16 text-gray-400"><i class="fas fa-search text-4xl mb-3"></i><p class="text-sm">No services match your criteria</p></div>';
    return;
  }

  let html = '<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">';
  filtered.forEach(s => {
    const svcId = s.id || s.ID;
    const name = s.name || s.Name || 'Service';
    const description = s.description || s.Description || '';
    const duration = s.duration || s.Duration || 0;
    const minPrice = parseFloat(s.min_price || s.MinPrice || 0);
    const maxPrice = parseFloat(s.max_price || s.MaxPrice || 0);
    const category = s.category || s.Category || '';
    const imgUrl = s.image_url || s.ImageURL || '';

    const priceDisplay = maxPrice > minPrice
      ? `$${minPrice.toFixed(2)} - $${maxPrice.toFixed(2)}`
      : minPrice > 0 ? `$${minPrice.toFixed(2)}` : 'Contact for price';

    html += `
      <div class="service-picker-item border-2 rounded-xl transition-all duration-200 bg-white overflow-hidden border-gray-100 hover:border-purple-200 hover:shadow-sm cursor-pointer"
        onclick="selectService(${svcId})">
        <div class="flex p-3 gap-3">
          <div class="w-16 h-16 rounded-lg flex-shrink-0 overflow-hidden bg-gradient-to-br from-purple-50 to-pink-50 border border-purple-100 flex items-center justify-center text-2xl">
            ${imgUrl ? `<img src="${imgUrl}" class="w-full h-full object-cover">` :
              `<i class="fas fa-concierge-bell text-purple-300"></i>`}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-1">
              <h5 class="font-semibold text-sm text-gray-900 truncate">${name}</h5>
              <span class="text-sm font-bold text-purple-700 flex-shrink-0 ml-1 whitespace-nowrap">${priceDisplay}</span>
            </div>
            ${category ? `<span class="text-[10px] text-purple-500 bg-purple-50 px-1.5 py-0.5 rounded inline-block mt-0.5">${category}</span>` : ''}
            ${description ? `<p class="text-xs text-gray-500 mt-1 line-clamp-2 leading-relaxed">${description}</p>` : ''}
            ${duration ? `<div class="flex items-center gap-1 mt-1.5 text-[11px] text-gray-400"><i class="far fa-clock"></i><span>${duration} min</span></div>` : ''}
          </div>
        </div>
        <div class="px-3 pb-3">
          <div class="w-full py-2 rounded-lg bg-gradient-to-r from-purple-500 to-pink-600 text-white text-xs font-semibold text-center hover:from-purple-600 hover:to-pink-700 transition active:scale-[0.98] shadow-sm">
            <i class="fas fa-plus mr-1"></i> Select Service
          </div>
        </div>
      </div>`;
  });
  html += '</div>';
  grid.innerHTML = html;
}

function filterPickerServices() {
  renderServicePicker(document.getElementById('servicePickerCategoryFilter')?.querySelector('.bg-purple-600.text-white')?.textContent || '');
}

function filterServiceCategory(category) {
  renderServicePicker(category);
  const searchInput = document.getElementById('servicePickerSearch');
  if (searchInput) searchInput.value = '';
}

// ========== SELECTION & STEP 2 ==========

function selectService(svcId) {
  const svc = servicePickerServices.find(s => (s.id || s.ID) === svcId);
  if (!svc) return;

  servicePickerSelectedService = {
    id: svcId,
    name: svc.name || svc.Name || 'Service',
    description: svc.description || svc.Description || '',
    duration: svc.duration || svc.Duration || 0,
    minPrice: parseFloat(svc.min_price || svc.MinPrice || 0),
    maxPrice: parseFloat(svc.max_price || svc.MaxPrice || 0),
    category: svc.category || svc.Category || '',
    imgUrl: svc.image_url || svc.ImageURL || ''
  };

  document.getElementById('servicePickerServiceId').value = svcId;
  document.getElementById('selectedServiceName').textContent = servicePickerSelectedService.name;
  document.getElementById('selectedServiceDuration').textContent = servicePickerSelectedService.duration ? `${servicePickerSelectedService.duration} min` : '';
  const price = servicePickerSelectedService.maxPrice > servicePickerSelectedService.minPrice
    ? `$${servicePickerSelectedService.minPrice.toFixed(2)} - $${servicePickerSelectedService.maxPrice.toFixed(2)}`
    : servicePickerSelectedService.minPrice > 0 ? `$${servicePickerSelectedService.minPrice.toFixed(2)}` : '';
  document.getElementById('selectedServicePrice').textContent = price;
  document.getElementById('selectedServiceDesc').textContent = servicePickerSelectedService.description;
  document.getElementById('selectedServiceCategory').textContent = servicePickerSelectedService.category || '';

  const imgContainer = document.getElementById('selectedServiceImage');
  if (imgContainer) {
    if (servicePickerSelectedService.imgUrl) {
      imgContainer.innerHTML = `<img src="${servicePickerSelectedService.imgUrl}" class="w-full h-full object-cover">`;
    } else {
      imgContainer.innerHTML = '<i class="fas fa-concierge-bell text-purple-600 text-2xl"></i>';
    }
  }

  const today = new Date();
  const dateInput = document.getElementById('servicePickerDate');
  const timeInput = document.getElementById('servicePickerTime');
  if (dateInput && !dateInput.value) dateInput.value = today.toISOString().split('T')[0];
  if (timeInput && !timeInput.value) {
    timeInput.value = `${String(today.getHours()).padStart(2, '0')}:${String(today.getMinutes()).padStart(2, '0')}`;
  }
  document.getElementById('servicePickerNotes').value = '';

  servicePickerGoToStep(2);
}

function backToServiceGrid() {
  servicePickerGoToStep(1);
}

// ========== STEP 3: CONFIRMATION ==========

function renderServiceConfirmation() {
  const container = document.getElementById('servicePickerConfirmInfo');
  if (!container || !servicePickerSelectedService) return;

  const date = document.getElementById('servicePickerDate').value;
  const time = document.getElementById('servicePickerTime').value;
  const notes = document.getElementById('servicePickerNotes').value;

  const dateObj = new Date(`${date}T${time}`);
  const dateStr = dateObj.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' });
  const timeStr = dateObj.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });

  const price = servicePickerSelectedService.maxPrice > servicePickerSelectedService.minPrice
    ? `$${servicePickerSelectedService.minPrice.toFixed(2)} - $${servicePickerSelectedService.maxPrice.toFixed(2)}`
    : servicePickerSelectedService.minPrice > 0 ? `$${servicePickerSelectedService.minPrice.toFixed(2)}` : '';

  container.innerHTML = `
    <div class="bg-gradient-to-br from-purple-50 to-pink-50 rounded-xl border border-purple-200 p-4 mb-4">
      <div class="flex items-center gap-3 mb-3">
        <div class="w-12 h-12 rounded-xl bg-purple-100 flex items-center justify-center text-xl flex-shrink-0 overflow-hidden">
          <i class="fas fa-concierge-bell text-purple-600"></i>
        </div>
        <div>
          <h4 class="font-bold text-gray-900">${servicePickerSelectedService.name}</h4>
          ${servicePickerSelectedService.category ? `<span class="text-xs text-purple-500">${servicePickerSelectedService.category}</span>` : ''}
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3 text-sm">
        <div class="bg-white/60 rounded-lg p-2.5">
          <p class="text-[10px] text-gray-400 uppercase tracking-wider">Date</p>
          <p class="font-semibold text-gray-800">${dateStr}</p>
        </div>
        <div class="bg-white/60 rounded-lg p-2.5">
          <p class="text-[10px] text-gray-400 uppercase tracking-wider">Time</p>
          <p class="font-semibold text-gray-800">${timeStr}</p>
        </div>
        ${servicePickerSelectedService.duration ? `
        <div class="bg-white/60 rounded-lg p-2.5">
          <p class="text-[10px] text-gray-400 uppercase tracking-wider">Duration</p>
          <p class="font-semibold text-gray-800">${servicePickerSelectedService.duration} min</p>
        </div>` : ''}
        ${price ? `
        <div class="bg-white/60 rounded-lg p-2.5">
          <p class="text-[10px] text-gray-400 uppercase tracking-wider">Estimated Price</p>
          <p class="font-semibold text-purple-700">${price}</p>
        </div>` : ''}
      </div>
      ${notes ? `<div class="mt-3 bg-white/60 rounded-lg p-2.5 text-sm text-gray-600"><span class="text-[10px] text-gray-400 uppercase tracking-wider block mb-0.5">Notes</span>${notes}</div>` : ''}
    </div>`;
}

// ========== SUBMIT ==========

async function submitServiceBooking() {
  if (!servicePickerSelectedService) return;

  const date = document.getElementById('servicePickerDate').value;
  const time = document.getElementById('servicePickerTime').value;
  const notes = document.getElementById('servicePickerNotes').value;

  if (!date || !time) {
    showNotification('Please select a date and time', 'error');
    return;
  }

  const [hours, minutes] = time.split(':');
  const bookingDateTime = `${date}T${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:00Z`;

  const submitBtn = document.getElementById('servicePickerSubmitBtn');
  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i> Submitting...';
  }

  try {
    let resp, data;

    if (servicePickerMode === 'client') {
      resp = await fetch('/client/bookings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          service_id: servicePickerSelectedService.id,
          scheduled_date: bookingDateTime,
          notes: notes,
          business_id: servicePickerBusinessId ? parseInt(servicePickerBusinessId) : parseInt(businessId)
        })
      });
      data = await resp.json();
      if (data.success) {
        hideServicePicker();
        showNotification('Booking request sent successfully!', 'success');
        if (typeof startMessagePolling === 'function') {
          setTimeout(() => {
            const bizId = servicePickerBusinessId || businessId;
            fetch(`/client/businesses/${bizId}/messages`)
              .then(r => r.text())
              .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newMsgs = doc.getElementById('messages-container');
                const curMsgs = document.getElementById('messages-container');
                if (newMsgs && curMsgs) curMsgs.innerHTML = newMsgs.innerHTML;
              }).catch(console.error);
          }, 500);
        }
      } else {
        showNotification(data.error || 'Failed to create booking', 'error');
      }
    } else {
      const clientId = typeof clientId !== 'undefined' ? clientId : servicePickerClientId;
      if (!clientId) {
        showNotification('Client ID not found', 'error');
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.innerHTML = servicePickerMode === 'client' ? '<i class="fas fa-paper-plane mr-1"></i> Send Request' : '<i class="fas fa-calendar-check mr-1"></i> Create Booking';
        }
        return;
      }
      resp = await fetch('/bookings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: parseInt(clientId),
          service_id: servicePickerSelectedService.id,
          customer_name: '',
          customer_email: '',
          customer_phone: '',
          booking_date: bookingDateTime,
          notes: notes
        })
      });
      data = await resp.json();
      if (data.success) {
        hideServicePicker();
        showNotification(data.message || 'Booking created successfully!', 'success');
        if (typeof fetchMessages === 'function') {
          setTimeout(() => fetchMessages(), 500);
        }
      } else {
        showNotification(data.error || 'Failed to create booking', 'error');
      }
    }
  } catch (e) {
    console.error('Failed to create booking:', e);
    showNotification('Failed to create booking', 'error');
  }

  if (submitBtn) {
    submitBtn.disabled = false;
    submitBtn.innerHTML = servicePickerMode === 'client'
      ? '<i class="fas fa-paper-plane mr-1"></i> Send Request'
      : '<i class="fas fa-calendar-check mr-1"></i> Create Booking';
  }
}
