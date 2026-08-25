const $ = (selector) => document.querySelector(selector);

const state = {
  items: [],
  tags: [],
  activeViewItem: null,
};

async function api(path, options = {}) {
  const headers = { ...(options.headers || {}) };
  if (options.body && !(options.body instanceof FormData) && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json";
  }

  const response = await fetch(path, { ...options, headers });

  if (!response.ok) {
    let message = `HTTP ${response.status}`;
    try {
      const body = await response.json();
      message = body.error || message;
    } catch (_) {}
    throw new Error(message);
  }

  if (response.status === 204) return null;
  return response.json();
}

function showToast(message) {
  const toast = $("#toast");
  toast.textContent = message;
  toast.classList.remove("hidden");
  toast.style.animation = "none";
  void toast.offsetWidth;
  toast.style.animation = "";
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.add("hidden"), 2400);
}

function normalizeTag(tag) {
  return {
    id: Number(tag.id ?? tag.ID),
    name: tag.name ?? tag.Name ?? "",
  };
}

function normalizeItem(item) {
  return {
    id: Number(item.id ?? item.ID),
    name: item.name ?? item.Name ?? "",
    short_description: item.short_description ?? item.ShortDescription ?? "",
    description: item.description ?? item.Description ?? "",
    image_url: item.image_url ?? item.ImageURL ?? "",
    price: Number(item.price ?? item.Price ?? 0),
    price_currency: item.price_currency ?? item.PriceCurrency ?? "RUB",
    usd_exchange_rate: Number(item.usd_exchange_rate ?? item.USDExchangeRate ?? 0),
    purchased_at: item.purchased_at ?? item.PurchasedAt ?? null,
    tags: (item.tags ?? item.Tags ?? []).map(normalizeTag),
    created_at: item.created_at ?? item.CreatedAt,
    updated_at: item.updated_at ?? item.UpdatedAt,
  };
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function formatMoney(value, currency) {
  return new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(Number(value || 0));
}

function convertedPrice(item) {
  if (!item.usd_exchange_rate) return "";
  if (item.price_currency === "USD") {
    return `≈ ${formatMoney(item.price * item.usd_exchange_rate, "RUB")}`;
  }
  return `≈ ${formatMoney(item.price / item.usd_exchange_rate, "USD")}`;
}

function formatDate(value) {
  if (!value) return "Не указана";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value).slice(0, 10);
  return new Intl.DateTimeFormat("ru-RU").format(date);
}

function dateInputValue(value) {
  if (!value) return "";
  return String(value).slice(0, 10);
}

function renderItemTags(tags, limit = 4) {
  const visible = tags.slice(0, limit);
  const rest = tags.length - visible.length;
  return [
    ...visible.map((tag) => `<span class="item-tag">${escapeHTML(tag.name)}</span>`),
    rest > 0 ? `<span class="item-tag">+${rest}</span>` : "",
  ].join("");
}

function renderItems() {
  const grid = $("#items-grid");
  const empty = $("#empty-state");

  grid.innerHTML = state.items.map((item, index) => `
    <article class="item-card" data-view-id="${item.id}" style="animation-delay:${Math.min(index * 35, 280)}ms">
      <div class="item-media">
        ${item.image_url
          ? `<img class="item-image" src="${escapeHTML(item.image_url)}" alt="${escapeHTML(item.name)}" loading="lazy" onerror="this.replaceWith(Object.assign(document.createElement('div'),{className:'item-placeholder',textContent:'📦'}))">`
          : `<div class="item-placeholder">📦</div>`}
      </div>
      <div class="item-body">
        <h3>${escapeHTML(item.name)}</h3>
        ${item.short_description ? `<p class="item-description">${escapeHTML(item.short_description)}</p>` : `<p class="item-description muted">Без краткого описания</p>`}
        <div class="item-tags">${renderItemTags(item.tags)}</div>
        ${item.purchased_at ? `<div class="item-meta">Приобретён: ${escapeHTML(formatDate(item.purchased_at))}</div>` : ""}
        <div class="item-price-block">
          <div class="item-price">${formatMoney(item.price, item.price_currency)}</div>
          <div class="item-price-equivalent">${escapeHTML(convertedPrice(item))}</div>
        </div>
        <div class="item-actions">
          <button class="small-button pressable" data-action="view" data-id="${item.id}">Подробнее</button>
          <button class="small-button pressable" data-action="edit" data-id="${item.id}">Изменить</button>
          <button class="small-button pressable" data-action="tags" data-id="${item.id}">Теги</button>
          <button class="danger small-button pressable" data-action="delete" data-id="${item.id}">Удалить</button>
        </div>
      </div>
    </article>
  `).join("");

  empty.classList.toggle("hidden", state.items.length !== 0);
  $("#items-status").textContent = `Показано: ${state.items.length}`;
}

function renderTags() {
  $("#tags-list").innerHTML = state.tags.map((tag) => `
    <span class="tag-chip">
      ${escapeHTML(tag.name)}
      <button class="pressable" title="Удалить тег" data-delete-tag="${tag.id}" data-tag-name="${escapeHTML(tag.name)}">×</button>
    </span>
  `).join("");

  $("#filter-tags").innerHTML = state.tags.map((tag) =>
    `<option value="${escapeHTML(tag.name)}">${escapeHTML(tag.name)}</option>`
  ).join("");
}

async function loadTags() {
  state.tags = (await api("/tags?limit=100")).map(normalizeTag);
  renderTags();
}

function buildItemsQuery() {
  const params = new URLSearchParams();
  const name = $("#filter-name").value.trim();
  const minPrice = $("#filter-min-price").value;
  const maxPrice = $("#filter-max-price").value;

  if (name) params.set("name", name);
  if (minPrice) params.set("min_price", minPrice);
  if (maxPrice) params.set("max_price", maxPrice);

  for (const option of $("#filter-tags").selectedOptions) {
    params.append("tag", option.value);
  }

  params.set("limit", "100");
  return params.toString();
}

async function loadItems() {
  try {
    $("#items-status").textContent = "Загрузка…";
    state.items = (await api(`/items?${buildItemsQuery()}`)).map(normalizeItem);
    renderItems();
  } catch (err) {
    $("#items-status").textContent = "Ошибка загрузки";
    showToast(err.message);
  }
}

function showImagePreview(url) {
  const wrap = $("#image-preview-wrap");
  const image = $("#image-preview");
  if (!url) {
    image.removeAttribute("src");
    wrap.classList.add("hidden");
    return;
  }
  image.src = url;
  wrap.classList.remove("hidden");
}

function resetItemForm() {
  $("#item-form").reset();
  $("#item-id").value = "";
  $("#item-image-url").value = "";
  $("#item-price").value = "0";
  $("#item-price-currency").value = "RUB";
  $("#item-usd-exchange-rate").value = "";
  $("#item-price-symbol").textContent = "₽";
  showImagePreview("");
}

function openCreateDialog() {
  resetItemForm();
  $("#item-dialog-title").textContent = "Новый предмет";
  $("#item-dialog").showModal();
  setTimeout(() => $("#item-name").focus(), 60);
}

function openEditDialog(item) {
  resetItemForm();
  $("#item-dialog-title").textContent = "Редактирование";
  $("#item-id").value = item.id;
  $("#item-name").value = item.name;
  $("#item-short-description").value = item.short_description;
  $("#item-description").value = item.description;
  $("#item-image-url").value = item.image_url;
  $("#item-price").value = item.price;
  $("#item-price-currency").value = item.price_currency;
  $("#item-price-symbol").textContent = item.price_currency === "USD" ? "$" : "₽";
  $("#item-purchased-at").value = dateInputValue(item.purchased_at);
  $("#item-usd-exchange-rate").value = item.usd_exchange_rate || "";
  showImagePreview(item.image_url);
  $("#item-dialog").showModal();
}

async function uploadImage(file) {
  const form = new FormData();
  form.append("image", file);
  return api("/uploads", { method: "POST", body: form });
}

function setButtonLoading(button, loading, text) {
  if (!button) return;
  if (loading) {
    button.dataset.originalText = button.textContent;
    button.textContent = text;
    button.classList.add("loading-dot");
    button.disabled = true;
  } else {
    button.textContent = button.dataset.originalText || "Сохранить";
    button.classList.remove("loading-dot");
    button.disabled = false;
  }
}

async function saveItem(event) {
  event.preventDefault();

  const saveButton = $("#save-item");
  setButtonLoading(saveButton, true, "Сохраняю");

  try {
    let imageURL = $("#item-image-url").value;
    const imageFile = $("#item-image-file").files[0];
    if (imageFile) {
      saveButton.textContent = "Загружаю фото";
      const upload = await uploadImage(imageFile);
      imageURL = upload.url;
      saveButton.textContent = "Сохраняю";
    }

    const id = $("#item-id").value;
    const payload = {
      name: $("#item-name").value.trim(),
      short_description: $("#item-short-description").value.trim(),
      description: $("#item-description").value.trim(),
      image_url: imageURL,
      price: Number($("#item-price").value || 0),
      price_currency: $("#item-price-currency").value,
      usd_exchange_rate: Number($("#item-usd-exchange-rate").value || 0),
      purchased_at: $("#item-purchased-at").value,
    };

    await api(id ? `/items/${id}` : "/items", {
      method: id ? "PUT" : "POST",
      body: JSON.stringify(payload),
    });

    $("#item-dialog").close();
    showToast(id ? "Предмет обновлён" : "Предмет создан");
    await loadItems();
  } catch (err) {
    showToast(err.message);
  } finally {
    setButtonLoading(saveButton, false);
  }
}

async function deleteItem(id) {
  if (!confirm("Удалить предмет?")) return;
  try {
    await api(`/items/${id}`, { method: "DELETE" });
    if ($("#view-dialog").open) $("#view-dialog").close();
    showToast("Предмет удалён");
    await loadItems();
  } catch (err) {
    showToast(err.message);
  }
}

function renderView(item) {
  state.activeViewItem = item;
  $("#view-content").innerHTML = `
    <div class="view-grid">
      <div class="view-media">
        ${item.image_url
          ? `<img src="${escapeHTML(item.image_url)}" alt="${escapeHTML(item.name)}">`
          : `<div class="view-placeholder">📦</div>`}
      </div>
      <div class="view-info">
        <h2>${escapeHTML(item.name)}</h2>
        <div class="view-price">${formatMoney(item.price, item.price_currency)}</div>
        <div class="view-equivalent">${escapeHTML(convertedPrice(item))}</div>
        ${item.short_description ? `<p class="view-summary">${escapeHTML(item.short_description)}</p>` : ""}
        <div class="view-description-block">
          <h3>Описание</h3>
          ${item.description ? `<p class="view-description">${escapeHTML(item.description)}</p>` : `<p class="view-description muted">Полное описание не указано.</p>`}
        </div>
        <div class="item-tags">${renderItemTags(item.tags, 100) || `<span class="muted small">Нет тегов</span>`}</div>
        <div class="view-details">
          <div class="detail"><span>Дата приобретения</span><strong>${escapeHTML(formatDate(item.purchased_at))}</strong></div>
          <div class="detail"><span>Курс доллара</span><strong>${item.usd_exchange_rate ? `${item.usd_exchange_rate.toLocaleString("ru-RU", { maximumFractionDigits: 4 })} ₽ за $1` : "Не указан"}</strong></div>
        </div>
        <div class="view-actions">
          <button class="primary pressable" data-view-action="edit" data-id="${item.id}">Изменить</button>
          <button class="secondary pressable" data-view-action="tags" data-id="${item.id}">Управлять тегами</button>
          <button class="danger pressable" data-view-action="delete" data-id="${item.id}">Удалить</button>
        </div>
      </div>
    </div>
  `;
}

async function openViewDialog(id) {
  try {
    const item = normalizeItem(await api(`/items/${id}`));
    renderView(item);
    $("#view-dialog").showModal();
  } catch (err) {
    showToast(err.message);
  }
}

function renderAssignTags(item) {
  const attached = new Set(item.tags.map((tag) => tag.id));
  $("#assign-tags-list").innerHTML = state.tags.map((tag) => {
    const isAttached = attached.has(tag.id);
    return `
      <div class="assign-tag-row ${isAttached ? "attached" : ""}">
        <span>${escapeHTML(tag.name)}</span>
        <button class="${isAttached ? "danger" : "primary"} small-button pressable"
                data-tag-action="${isAttached ? "remove" : "add"}"
                data-tag-id="${tag.id}">
          ${isAttached ? "Убрать" : "Добавить"}
        </button>
      </div>
    `;
  }).join("") || `<p class="muted">Сначала создайте хотя бы один тег.</p>`;
}

async function openTagsDialog(item) {
  try {
    const freshItem = normalizeItem(await api(`/items/${item.id}`));
    $("#tags-item-id").value = freshItem.id;
    $("#tags-item-title").textContent = freshItem.name;
    renderAssignTags(freshItem);
    $("#tags-dialog").showModal();
  } catch (err) {
    showToast(err.message);
  }
}

async function changeItemTag(action, tagID, button) {
  const itemID = $("#tags-item-id").value;
  button.disabled = true;
  try {
    await api(`/items/${itemID}/tags/${tagID}`, {
      method: action === "add" ? "POST" : "DELETE",
    });
    showToast(action === "add" ? "Тег добавлен" : "Тег удалён");
    await loadItems();
    const freshItem = normalizeItem(await api(`/items/${itemID}`));
    renderAssignTags(freshItem);
    if (state.activeViewItem?.id === freshItem.id) renderView(freshItem);
  } catch (err) {
    showToast(err.message);
  } finally {
    button.disabled = false;
  }
}

async function createTag(event) {
  event.preventDefault();
  const input = $("#tag-name");
  const name = input.value.trim();
  if (!name) return;

  try {
    await api("/tags", { method: "POST", body: JSON.stringify({ name }) });
    input.value = "";
    showToast("Тег создан");
    await loadTags();
  } catch (err) {
    showToast(err.message);
  }
}

async function deleteTag(name) {
  if (!confirm(`Удалить тег «${name}»?`)) return;
  try {
    await api(`/tags/${encodeURIComponent(name)}`, { method: "DELETE" });
    showToast("Тег удалён");
    await Promise.all([loadTags(), loadItems()]);
  } catch (err) {
    showToast(err.message);
  }
}

function addRipple(event) {
  const button = event.target.closest(".pressable");
  if (!button || button.disabled) return;
  const rect = button.getBoundingClientRect();
  const dot = document.createElement("span");
  dot.className = "ripple-dot";
  dot.style.left = `${event.clientX - rect.left}px`;
  dot.style.top = `${event.clientY - rect.top}px`;
  button.append(dot);
  setTimeout(() => dot.remove(), 520);
}

document.addEventListener("pointerdown", addRipple);

$("#open-item-form").addEventListener("click", openCreateDialog);
$("#close-item-form").addEventListener("click", () => $("#item-dialog").close());
$("#cancel-item").addEventListener("click", () => $("#item-dialog").close());
$("#close-view-dialog").addEventListener("click", () => $("#view-dialog").close());
$("#close-tags-dialog").addEventListener("click", () => $("#tags-dialog").close());
$("#item-form").addEventListener("submit", saveItem);
$("#tag-form").addEventListener("submit", createTag);
$("#filters-form").addEventListener("submit", (event) => { event.preventDefault(); loadItems(); });
$("#reset-filters").addEventListener("click", () => { $("#filters-form").reset(); loadItems(); });

$("#item-price-currency").addEventListener("change", (event) => {
  $("#item-price-symbol").textContent = event.target.value === "USD" ? "$" : "₽";
});

$("#item-image-file").addEventListener("change", (event) => {
  const file = event.target.files[0];
  if (!file) {
    showImagePreview($("#item-image-url").value);
    return;
  }
  showImagePreview(URL.createObjectURL(file));
});

$("#remove-image").addEventListener("click", () => {
  $("#item-image-file").value = "";
  $("#item-image-url").value = "";
  showImagePreview("");
});

$("#items-grid").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-action]");
  if (button) {
    const item = state.items.find((value) => value.id === Number(button.dataset.id));
    if (!item) return;
    event.stopPropagation();

    if (button.dataset.action === "view") openViewDialog(item.id);
    if (button.dataset.action === "edit") openEditDialog(item);
    if (button.dataset.action === "tags") openTagsDialog(item);
    if (button.dataset.action === "delete") deleteItem(item.id);
    return;
  }

  const card = event.target.closest("[data-view-id]");
  if (card) openViewDialog(Number(card.dataset.viewId));
});

$("#view-content").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-view-action]");
  if (!button || !state.activeViewItem) return;

  const item = state.activeViewItem;
  if (button.dataset.viewAction === "edit") {
    $("#view-dialog").close();
    openEditDialog(item);
  }
  if (button.dataset.viewAction === "tags") {
    $("#view-dialog").close();
    openTagsDialog(item);
  }
  if (button.dataset.viewAction === "delete") deleteItem(item.id);
});

$("#tags-list").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-delete-tag]");
  if (!button) return;
  deleteTag(button.dataset.tagName);
});

$("#assign-tags-list").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-tag-action]");
  if (!button) return;
  changeItemTag(button.dataset.tagAction, button.dataset.tagId, button);
});

Promise.all([loadTags(), loadItems()]).catch((err) => showToast(err.message));
