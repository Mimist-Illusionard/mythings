const $ = (selector) => document.querySelector(selector);

const state = {
  items: [],
  tags: [],
};

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });

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
  clearTimeout(showToast.timer);
  showToast.timer = setTimeout(() => toast.classList.add("hidden"), 2200);
}

function normalizeItem(item) {
  return {
    id: item.id ?? item.ID,
    name: item.name ?? item.Name ?? "",
    description: item.description ?? item.Description ?? "",
    image_url: item.image_url ?? item.ImageURL ?? "",
    price: item.price ?? item.Price ?? 0,
    attributes: item.attributes ?? item.Attributes ?? {},
    created_at: item.created_at ?? item.CreatedAt,
    updated_at: item.updated_at ?? item.UpdatedAt,
  };
}

function normalizeTag(tag) {
  return {
    id: tag.id ?? tag.ID,
    name: tag.name ?? tag.Name ?? "",
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

function renderAttributes(attributes) {
  if (!attributes || typeof attributes !== "object") return "";
  return Object.entries(attributes)
    .slice(0, 5)
    .map(([key, value]) => `<span class="attribute">${escapeHTML(key)}: ${escapeHTML(value)}</span>`)
    .join("");
}

function renderItems() {
  const grid = $("#items-grid");
  const empty = $("#empty-state");

  grid.innerHTML = state.items.map((item) => `
    <article class="item-card">
      ${item.image_url
        ? `<img class="item-image" src="${escapeHTML(item.image_url)}" alt="${escapeHTML(item.name)}" onerror="this.replaceWith(Object.assign(document.createElement('div'),{className:'item-placeholder',textContent:'📦'}))">`
        : `<div class="item-placeholder">📦</div>`}
      <div class="item-body">
        <h3>${escapeHTML(item.name)}</h3>
        ${item.description ? `<p class="item-description">${escapeHTML(item.description)}</p>` : ""}
        <div class="attributes">${renderAttributes(item.attributes)}</div>
        <div class="item-price">${Number(item.price).toLocaleString("ru-RU")}</div>
        <div class="item-actions">
          <button class="small-button" data-action="edit" data-id="${item.id}">Изменить</button>
          <button class="small-button" data-action="tags" data-id="${item.id}">Теги</button>
          <button class="danger small-button" data-action="delete" data-id="${item.id}">Удалить</button>
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
      <button title="Удалить тег" data-delete-tag="${tag.id}" data-tag-name="${escapeHTML(tag.name)}">×</button>
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
    state.items = (await api(`/items?${buildItemsQuery()}`)).map(normalizeItem);
    renderItems();
  } catch (err) {
    showToast(err.message);
  }
}

function openCreateDialog() {
  $("#item-dialog-title").textContent = "Новый предмет";
  $("#item-id").value = "";
  $("#item-name").value = "";
  $("#item-description").value = "";
  $("#item-image-url").value = "";
  $("#item-price").value = "0";
  $("#item-attributes").value = "{}";
  $("#item-dialog").showModal();
}

function openEditDialog(item) {
  $("#item-dialog-title").textContent = "Редактирование";
  $("#item-id").value = item.id;
  $("#item-name").value = item.name;
  $("#item-description").value = item.description;
  $("#item-image-url").value = item.image_url;
  $("#item-price").value = item.price;
  $("#item-attributes").value = JSON.stringify(item.attributes || {}, null, 2);
  $("#item-dialog").showModal();
}

async function saveItem(event) {
  event.preventDefault();

  let attributes;
  try {
    attributes = JSON.parse($("#item-attributes").value || "{}");
  } catch (_) {
    showToast("Атрибуты должны быть валидным JSON");
    return;
  }

  const id = $("#item-id").value;
  const payload = {
    name: $("#item-name").value.trim(),
    description: $("#item-description").value.trim(),
    image_url: $("#item-image-url").value.trim(),
    price: Number($("#item-price").value || 0),
    attributes,
  };

  try {
    await api(id ? `/items/${id}` : "/items", {
      method: id ? "PUT" : "POST",
      body: JSON.stringify(payload),
    });
    $("#item-dialog").close();
    showToast(id ? "Предмет обновлён" : "Предмет создан");
    await loadItems();
  } catch (err) {
    showToast(err.message);
  }
}

async function deleteItem(id) {
  if (!confirm("Удалить предмет?")) return;
  try {
    await api(`/items/${id}`, { method: "DELETE" });
    showToast("Предмет удалён");
    await loadItems();
  } catch (err) {
    showToast(err.message);
  }
}

function openTagsDialog(item) {
  $("#tags-item-id").value = item.id;
  $("#tags-item-title").textContent = item.name;
  $("#assign-tags-list").innerHTML = state.tags.map((tag) => `
    <div class="assign-tag-row">
      <span>${escapeHTML(tag.name)} <span class="muted">#${tag.id}</span></span>
      <div>
        <button class="small-button" data-tag-action="add" data-tag-id="${tag.id}">Добавить</button>
        <button class="danger small-button" data-tag-action="remove" data-tag-id="${tag.id}">Удалить</button>
      </div>
    </div>
  `).join("");
  $("#tags-dialog").showModal();
}

async function changeItemTag(action, tagID) {
  const itemID = $("#tags-item-id").value;
  try {
    await api(`/items/${itemID}/tags/${tagID}`, {
      method: action === "add" ? "POST" : "DELETE",
    });
    showToast(action === "add" ? "Тег добавлен" : "Тег удалён");
  } catch (err) {
    showToast(err.message);
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
    await loadTags();
    await loadItems();
  } catch (err) {
    showToast(err.message);
  }
}

$("#open-item-form").addEventListener("click", openCreateDialog);
$("#close-item-form").addEventListener("click", () => $("#item-dialog").close());
$("#cancel-item").addEventListener("click", () => $("#item-dialog").close());
$("#item-form").addEventListener("submit", saveItem);
$("#tag-form").addEventListener("submit", createTag);
$("#filters-form").addEventListener("submit", (event) => { event.preventDefault(); loadItems(); });
$("#reset-filters").addEventListener("click", () => { $("#filters-form").reset(); loadItems(); });
$("#close-tags-dialog").addEventListener("click", () => $("#tags-dialog").close());

$("#items-grid").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const item = state.items.find((value) => value.id === Number(button.dataset.id));
  if (!item) return;

  if (button.dataset.action === "edit") openEditDialog(item);
  if (button.dataset.action === "tags") openTagsDialog(item);
  if (button.dataset.action === "delete") deleteItem(item.id);
});

$("#tags-list").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-delete-tag]");
  if (!button) return;
  deleteTag(button.dataset.tagName);
});

$("#assign-tags-list").addEventListener("click", (event) => {
  const button = event.target.closest("button[data-tag-action]");
  if (!button) return;
  changeItemTag(button.dataset.tagAction, button.dataset.tagId);
});

Promise.all([loadTags(), loadItems()]).catch((err) => showToast(err.message));
