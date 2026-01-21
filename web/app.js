const STORAGE_KEY = "simplebank_access_token";
const LAST_ACTION_KEY = "simplebank_last_action";

function $(id) { return document.getElementById(id); }
function has(id) { return !!$(id); }

function getToken() { return localStorage.getItem(STORAGE_KEY) || ""; }
function setToken(token) {
  if (!token) localStorage.removeItem(STORAGE_KEY);
  else localStorage.setItem(STORAGE_KEY, token);
}

function setLastAction(text) {
  if (!text) return;
  localStorage.setItem(LAST_ACTION_KEY, text);
}
function getLastAction() {
  return localStorage.getItem(LAST_ACTION_KEY) || "";
}

function showBanner(type, message) {
  const b = $("banner");
  if (!b) return;
  b.hidden = false;
  b.className = `banner ${type || ""}`.trim();
  b.textContent = message || "";
}
function clearBanner() {
  const b = $("banner");
  if (!b) return;
  b.hidden = true;
  b.textContent = "";
  b.className = "banner";
}

let toastTimer;
function toast(message) {
  const t = $("toast");
  if (!t) return;
  t.hidden = false;
  t.textContent = message;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    t.hidden = true;
    t.textContent = "";
  }, 2400);
}

async function apiFetch(path, { method = "GET", body } = {}) {
  const headers = { "Content-Type": "application/json" };
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`/api${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();
  let data;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }

  if (!res.ok) {
    const msg =
      (data && data.error) ||
      (data && data.message) ||
      (typeof data === "string" ? data : "") ||
      res.statusText ||
      "Request failed";
    const err = new Error(msg);
    err.status = res.status;
    err.body = data;
    throw err;
  }
  return data;
}

function currencyFmt(amount, currency) {
  const n = Number(amount);
  if (!Number.isFinite(n)) return String(amount);
  const cur = String(currency || "").toUpperCase();
  const symbol = cur === "INR" ? "₹" : cur === "USD" ? "$" : cur === "EUR" ? "€" : "";
  const abs = Math.abs(n);
  const formatted = abs.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  const sign = n < 0 ? "-" : "";
  return symbol ? `${sign}${symbol}${formatted}` : `${sign}${formatted} ${cur}`.trim();
}

function fxRate(from, to) {
  const inr = { INR: 1.0, USD: 83.0, EUR: 90.0 };
  const f = inr[String(from || "").toUpperCase()];
  const t = inr[String(to || "").toUpperCase()];
  if (!f || !t) return null;
  return f / t;
}

function fxConvert(amount, from, to) {
  const rate = fxRate(from, to);
  if (!rate) return null;
  const n = Number(amount);
  if (!Number.isFinite(n)) return null;
  return { rate, toAmount: Math.round(n * rate) };
}

function requireAuth() {
  const auth = document.body?.dataset?.auth;
  if (auth !== "protected") return;
  if (!getToken()) window.location.href = "/login.html";
}

function bindLogout() {
  if (!has("btnLogout")) return;
  $("btnLogout").addEventListener("click", () => {
    setToken("");
    toast("Signed out");
    window.location.href = "/login.html";
  });
}

async function loadAccounts({ pageId = 1, pageSize = 10 } = {}) {
  return await apiFetch(`/accounts?page_id=${pageId}&page_size=${pageSize}`);
}

function renderAccountsTable(tbody, accounts, { onView } = {}) {
  tbody.innerHTML = "";
  if (!accounts || accounts.length === 0) {
    tbody.innerHTML = `<tr><td colspan="4" class="muted">No accounts found.</td></tr>`;
    return;
  }
  for (const a of accounts) {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${a.id}</td>
      <td>${a.currency}</td>
      <td class="right">${currencyFmt(a.balance, a.currency)}</td>
      <td class="right"><button class="btn btn-secondary" data-view="${a.id}">View</button></td>
    `;
    tbody.appendChild(tr);
  }
  if (onView) {
    tbody.querySelectorAll("[data-view]").forEach((btn) => {
      btn.addEventListener("click", () => onView(Number(btn.getAttribute("data-view"))));
    });
  }
}

async function initLanding() {
  if (getToken()) window.location.href = "/dashboard.html";
}

async function initLogin() {
  if (getToken()) window.location.href = "/dashboard.html";

  if (!has("loginForm")) return;
  $("loginForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    clearBanner();
    const fd = new FormData(e.target);
    const body = Object.fromEntries(fd.entries());
    try {
      const rsp = await apiFetch("/users/login", { method: "POST", body });
      if (rsp?.access_token) setToken(rsp.access_token);
      setLastAction("Signed in");
      toast("Signed in");
      window.location.href = "/dashboard.html";
    } catch (err) {
      showBanner("error", err.message || "Sign in failed");
    }
  });
}

async function initSignup() {
  if (!has("signupForm")) return;
  $("signupForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    clearBanner();
    const fd = new FormData(e.target);
    const body = Object.fromEntries(fd.entries());
    try {
      await apiFetch("/users", { method: "POST", body });
      setLastAction("Profile created");
      toast("Profile created. Please sign in.");
      window.location.href = "/login.html";
    } catch (err) {
      showBanner("error", err.message || "Signup failed");
    }
  });
}

async function initDashboard() {
  bindLogout();

  if (has("lastAction")) {
    $("lastAction").textContent = getLastAction() || "—";
  }

  const tbody = $("accountsTableBody");
  const refreshBtn = $("btnRefreshAccounts");
  const load = async () => {
    try {
      clearBanner();
      tbody.innerHTML = `<tr><td colspan="3" class="muted">Loading…</td></tr>`;
      const accounts = await loadAccounts({ pageId: 1, pageSize: 10 });
      const count = accounts.length;
      const totalsByCurrency = new Map();
      for (const a of accounts) {
        const cur = String(a.currency || "").toUpperCase();
        const bal = Number(a.balance || 0);
        totalsByCurrency.set(cur, (totalsByCurrency.get(cur) || 0) + bal);
      }

      if (has("accountsCount")) $("accountsCount").textContent = String(count);

      // Prefer showing INR total with ₹ on the dashboard when available.
      const inrTotal = totalsByCurrency.get("INR");
      if (has("totalBalance")) {
        if (typeof inrTotal === "number") $("totalBalance").textContent = currencyFmt(inrTotal, "INR");
        else if (totalsByCurrency.size === 1) {
          const [cur, val] = Array.from(totalsByCurrency.entries())[0];
          $("totalBalance").textContent = currencyFmt(val, cur);
        } else {
          $("totalBalance").textContent = "₹—";
        }
      }
      if (has("totalBreakdown")) {
        if (totalsByCurrency.size === 0) $("totalBreakdown").textContent = "Across all accounts";
        else if (totalsByCurrency.size === 1) $("totalBreakdown").textContent = "Across all accounts";
        else {
          const parts = Array.from(totalsByCurrency.entries())
            .sort(([a], [b]) => a.localeCompare(b))
            .map(([cur, val]) => currencyFmt(val, cur));
          $("totalBreakdown").textContent = `Totals: ${parts.join(" • ")}`;
        }
      }

      tbody.innerHTML = "";
      if (count === 0) {
        tbody.innerHTML = `<tr><td colspan="3" class="muted">No accounts found.</td></tr>`;
        return;
      }
      for (const a of accounts) {
        const tr = document.createElement("tr");
        tr.innerHTML = `
          <td>${a.id}</td>
          <td>${a.currency}</td>
          <td class="right">${currencyFmt(a.balance, a.currency)}</td>
        `;
        tbody.appendChild(tr);
      }
    } catch (err) {
      showBanner("error", err.message || "Failed to load accounts");
      tbody.innerHTML = `<tr><td colspan="3" class="muted">—</td></tr>`;
    }
  };

  if (refreshBtn) refreshBtn.addEventListener("click", load);
  await load();
}

async function initAccounts() {
  bindLogout();

  const tbody = $("accountsTableBody");
  const refreshBtn = $("btnRefreshAccounts");
  const details = $("accountDetails");
  const depositAccountId = $("depositAccountId");

  const viewAccount = async (id) => {
    try {
      clearBanner();
      const acc = await apiFetch(`/accounts/${id}`);
      details.innerHTML = `
        <div><b>Account ID</b>: ${acc.id}</div>
        <div><b>Currency</b>: ${acc.currency}</div>
        <div><b>Balance</b>: ${currencyFmt(acc.balance, acc.currency)}</div>
      `;
      if (depositAccountId) depositAccountId.value = String(acc.id);
    } catch (err) {
      showBanner("error", err.message || "Failed to load account details");
    }
  };

  const load = async () => {
    try {
      clearBanner();
      tbody.innerHTML = `<tr><td colspan="4" class="muted">Loading…</td></tr>`;
      const accounts = await loadAccounts({ pageId: 1, pageSize: 10 });
      renderAccountsTable(tbody, accounts, { onView: viewAccount });
    } catch (err) {
      showBanner("error", err.message || "Failed to load accounts");
      tbody.innerHTML = `<tr><td colspan="4" class="muted">—</td></tr>`;
    }
  };

  if (refreshBtn) refreshBtn.addEventListener("click", load);

  if (has("openAccountForm")) {
    $("openAccountForm").addEventListener("submit", async (e) => {
      e.preventDefault();
      clearBanner();
      const fd = new FormData(e.target);
      const body = Object.fromEntries(fd.entries());
      try {
        await apiFetch("/accounts", { method: "POST", body });
        setLastAction("Account opened");
        toast("Account opened");
        await load();
      } catch (err) {
        showBanner("error", err.message || "Failed to open account");
      }
    });
  }

  if (has("depositForm")) {
    $("depositForm").addEventListener("submit", async (e) => {
      e.preventDefault();
      clearBanner();
      const fd = new FormData(e.target);
      const body = Object.fromEntries(fd.entries());
      const id = Number(body.id);
      const amount = Number(body.amount);
      if (!id || !amount || amount <= 0) {
        showBanner("error", "Enter a valid account ID and amount.");
        return;
      }
      try {
        await apiFetch(`/accounts/${id}/deposit`, { method: "POST", body: { amount } });
        setLastAction("Money added");
        toast("Money added");
        await load();
        await viewAccount(id);
        e.target.reset();
        if (depositAccountId) depositAccountId.value = String(id);
      } catch (err) {
        showBanner("error", err.message || "Failed to add money");
      }
    });
  }

  await load();
}

async function initTransfers() {
  bindLogout();

  const fromSelect = $("fromAccount");
  const toSelfSelect = $("toAccountSelf");
  const currencyInput = $("transferCurrency");
  const hint = $("fromAccountHint");
  const toSelfHint = $("toAccountSelfHint");
  const summary = $("transferSummary");
  const refreshBtn = $("btnRefreshAccounts");
  const modeSelf = $("modeSelf");
  const modeOther = $("modeOther");
  const selfFields = $("selfFields");
  const otherFields = $("otherFields");
  const recipientUser = $("recipientUsername");
  const recipientId = $("toAccountId");
  const recipientLookup = $("recipientLookup");

  let accounts = [];
  let transferMode = "self"; // "self" | "other"
  let recipientMeta = null; // {id, owner, currency}

  const transfersTbody = $("transfersTableBody");
  const refreshTransfersBtn = $("btnRefreshTransfers");

  const fmtIST = (iso) => {
    try {
      return new Date(iso).toLocaleString("en-IN", {
        timeZone: "Asia/Kolkata",
        year: "numeric",
        month: "short",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return String(iso || "");
    }
  };

  const renderTransfers = (items) => {
    if (!transfersTbody) return;
    transfersTbody.innerHTML = "";
    if (!items || items.length === 0) {
      transfersTbody.innerHTML = `<tr><td colspan="5" class="muted">No transfers yet.</td></tr>`;
      return;
    }

    // Build set of owned account ids from loaded accounts
    const owned = new Set(accounts.map((a) => Number(a.id)));

    for (const t of items) {
      const fromOwned = owned.has(Number(t.from_account_id));
      const toOwned = owned.has(Number(t.to_account_id));
      const type = fromOwned && toOwned ? "Self" : fromOwned ? "Sent" : "Received";

      let amountText = currencyFmt(t.amount, t.from_currency || "");
      if (t.from_currency && t.to_currency && t.from_currency !== t.to_currency) {
        const fx = fxConvert(t.amount, t.from_currency, t.to_currency);
        if (fx) amountText = `${currencyFmt(t.amount, t.from_currency)} → ${currencyFmt(fx.toAmount, t.to_currency)}`;
      }

      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${fmtIST(t.created_at)}</td>
        <td>${type}</td>
        <td>Acct ${t.from_account_id} <span class="muted small">${t.from_currency || ""}</span></td>
        <td>Acct ${t.to_account_id} <span class="muted small">${t.to_currency || ""}</span></td>
        <td class="right">${amountText}</td>
      `;
      transfersTbody.appendChild(tr);
    }
  };

  const loadTransfers = async () => {
    if (!transfersTbody) return;
    try {
      transfersTbody.innerHTML = `<tr><td colspan="5" class="muted">Loading…</td></tr>`;
      const items = await apiFetch("/transfers?page_id=1&page_size=10");
      renderTransfers(items);
    } catch (err) {
      transfersTbody.innerHTML = `<tr><td colspan="5" class="muted">Failed to load transfers.</td></tr>`;
    }
  };

  const updateFrom = () => {
    const id = Number(fromSelect.value);
    const acc = accounts.find((a) => Number(a.id) === id);
    if (!acc) {
      currencyInput.value = "";
      if (hint) hint.textContent = "";
      return;
    }
    currencyInput.value = acc.currency;
    if (hint) hint.textContent = `Available balance: ${currencyFmt(acc.balance, acc.currency)}`;
  };

  const selectedToCurrency = () => {
    if (transferMode === "self") {
      const toId = Number(toSelfSelect.value);
      const toAcc = accounts.find((a) => Number(a.id) === toId);
      return toAcc ? toAcc.currency : "";
    }
    return recipientMeta ? recipientMeta.currency : "";
  };

  const renderSummary = () => {
    const fromId = fromSelect?.value;
    const toId = transferMode === "self" ? toSelfSelect?.value : recipientId?.value;
    const amt = $("transferAmount")?.value;
    const fromCur = currencyInput?.value;
    const toCur = selectedToCurrency();
    if (!summary) return;
    if (!fromId || !toId || !amt || !fromCur) {
      summary.textContent = "Fill in the form to see transfer details.";
      return;
    }

    let extra = "";
    if (toCur && fromCur && toCur !== fromCur) {
      const fx = fxConvert(amt, fromCur, toCur);
      if (fx) {
        extra = `<div class="muted small">FX rate: ${fx.rate.toFixed(4)} • Recipient gets: <b>${currencyFmt(fx.toAmount, toCur)}</b></div>`;
      }
    }

    if (transferMode === "other") {
      const owner = recipientMeta?.owner ? ` (${recipientMeta.owner})` : "";
      summary.innerHTML = `
        <div><b>From</b>: Account ${fromId}</div>
        <div><b>To</b>: Account ${toId}${owner}</div>
        <div><b>Send</b>: ${currencyFmt(amt, fromCur)}</div>
        ${extra}
      `;
      return;
    }

    summary.innerHTML = `
      <div><b>From</b>: Account ${fromId}</div>
      <div><b>To</b>: Account ${toId}</div>
      <div><b>Send</b>: ${currencyFmt(amt, fromCur)}</div>
      ${extra}
    `;
  };

  const load = async () => {
    try {
      clearBanner();
      fromSelect.innerHTML = `<option value="">Loading…</option>`;
      accounts = await loadAccounts({ pageId: 1, pageSize: 10 });
      fromSelect.innerHTML = `<option value="">Select an account</option>`;
      for (const a of accounts) {
        const opt = document.createElement("option");
        opt.value = String(a.id);
        opt.textContent = `Account ${a.id} • ${a.currency} • ${currencyFmt(a.balance, a.currency)}`;
        fromSelect.appendChild(opt);
      }

      if (toSelfSelect) {
        toSelfSelect.innerHTML = `<option value="">Select an account</option>`;
        for (const a of accounts) {
          const opt = document.createElement("option");
          opt.value = String(a.id);
          opt.textContent = `Account ${a.id} • ${a.currency} • ${currencyFmt(a.balance, a.currency)}`;
          toSelfSelect.appendChild(opt);
        }
      }

      updateFrom();
      if (toSelfHint) toSelfHint.textContent = "";
      renderSummary();
      await loadTransfers();
    } catch (err) {
      showBanner("error", err.message || "Failed to load accounts");
      fromSelect.innerHTML = `<option value="">—</option>`;
      if (toSelfSelect) toSelfSelect.innerHTML = `<option value="">—</option>`;
    }
  };

  const setMode = (mode) => {
    transferMode = mode;
    recipientMeta = null;
    if (recipientLookup) recipientLookup.textContent = "";
    if (modeSelf && modeOther) {
      modeSelf.classList.toggle("active", mode === "self");
      modeOther.classList.toggle("active", mode === "other");
      modeSelf.setAttribute("aria-selected", mode === "self" ? "true" : "false");
      modeOther.setAttribute("aria-selected", mode === "other" ? "true" : "false");
    }
    if (selfFields) selfFields.hidden = mode !== "self";
    if (otherFields) otherFields.hidden = mode !== "other";
    // required toggles
    if (toSelfSelect) toSelfSelect.required = mode === "self";
    if (recipientId) recipientId.required = mode === "other";
    renderSummary();
  };

  if (modeSelf) modeSelf.addEventListener("click", () => setMode("self"));
  if (modeOther) modeOther.addEventListener("click", () => setMode("other"));

  if (fromSelect) fromSelect.addEventListener("change", () => {
    updateFrom();
    // If self mode and from==to, clear to selection
    if (transferMode === "self" && toSelfSelect && toSelfSelect.value === fromSelect.value) {
      toSelfSelect.value = "";
    }
    renderSummary();
  });

  if (toSelfSelect) {
    toSelfSelect.addEventListener("change", () => {
      const toId = Number(toSelfSelect.value);
      const toAcc = accounts.find((a) => Number(a.id) === toId);
      if (toSelfHint) {
        toSelfHint.textContent = toAcc ? `Currency: ${toAcc.currency}` : "";
      }
      // prevent selecting same account
      if (toSelfSelect.value && toSelfSelect.value === fromSelect.value) {
        toSelfSelect.value = "";
        if (toSelfHint) toSelfHint.textContent = "Choose a different account.";
      }
      renderSummary();
    });
  }

  if (recipientId) {
    const doLookup = async () => {
      const id = Number(recipientId.value);
      recipientMeta = null;
      if (recipientLookup) recipientLookup.textContent = "";
      if (!id) return;
      try {
        const meta = await apiFetch(`/accounts/${id}/lookup`);
        recipientMeta = meta;
        const nameOk = recipientUser?.value ? (meta.owner === recipientUser.value) : true;
        if (recipientLookup) {
          recipientLookup.textContent = nameOk
            ? `Recipient: ${meta.owner} • Currency: ${meta.currency}`
            : `Account belongs to ${meta.owner} (username mismatch)`;
        }
      } catch (err) {
        if (recipientLookup) recipientLookup.textContent = "Account not found.";
      } finally {
        renderSummary();
      }
    };
    recipientId.addEventListener("blur", doLookup);
    recipientId.addEventListener("input", () => { recipientMeta = null; renderSummary(); });
    if (recipientUser) recipientUser.addEventListener("input", () => { renderSummary(); });
  }

  if (has("transferAmount")) $("transferAmount").addEventListener("input", renderSummary);
  if (refreshBtn) refreshBtn.addEventListener("click", load);
  if (refreshTransfersBtn) refreshTransfersBtn.addEventListener("click", loadTransfers);

  if (has("transferForm")) {
    $("transferForm").addEventListener("submit", async (e) => {
      e.preventDefault();
      clearBanner();
      const fd = new FormData(e.target);
      const body = Object.fromEntries(fd.entries());
      const fromId = Number(body.from_account_id);
      const amount = Number(body.amount);

      let toId;
      let toUsername = "";
      if (transferMode === "self") {
        toId = Number(toSelfSelect.value);
      } else {
        toId = Number(body.to_account_id);
        toUsername = (body.to_username || "").trim();
      }

      if (!fromId || !toId || !amount) {
        showBanner("error", "Please fill in all fields.");
        return;
      }
      if (fromId === toId) {
        showBanner("error", "From and To accounts must be different.");
        return;
      }

      // If transferring to others and a username was provided, require lookup match.
      if (transferMode === "other" && toUsername) {
        if (!recipientMeta || Number(recipientMeta.id) !== toId) {
          showBanner("error", "Please verify the recipient account ID (leave the field to validate).");
          return;
        }
        if (recipientMeta.owner !== toUsername) {
          showBanner("error", "Recipient username does not match the account owner.");
          return;
        }
      }
      try {
        const payload = {
          from_account_id: fromId,
          to_account_id: toId,
          amount: Math.round(amount),
        };
        if (transferMode === "other" && toUsername) payload.to_username = toUsername;
        await apiFetch("/transfers", { method: "POST", body: payload });
        setLastAction("Transfer sent");
        toast("Transfer sent");
        e.target.reset();
        if (hint) hint.textContent = "";
        if (toSelfHint) toSelfHint.textContent = "";
        if (recipientLookup) recipientLookup.textContent = "";
        recipientMeta = null;
        summary.textContent = "Transfer submitted successfully.";
        await load();
        await loadTransfers();
      } catch (err) {
        showBanner("error", err.message || "Transfer failed");
      }
    });
  }

  setMode("self");
  await load();
}

async function init() {
  requireAuth();

  const page = document.body?.dataset?.page || "";
  if (page === "landing") return initLanding();
  if (page === "login") return initLogin();
  if (page === "signup") return initSignup();
  if (page === "dashboard") return initDashboard();
  if (page === "accounts") return initAccounts();
  if (page === "transfers") return initTransfers();
}

document.addEventListener("DOMContentLoaded", init);


