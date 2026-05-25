const COLORS = {
  starlink: "#38bdf8",
  oneweb: "#8b5cf6",
  viasat: "#f59e0b",
  inmarsat: "#f97316",
  thuraya: "#fb7185",
  ses_o3b: "#14b8a6",
  hughes: "#ef4444",
  intelsat: "#a3e635",
  avanti: "#22c55e",
  eutelsat_skylogic: "#06b6d4",
  marlink: "#eab308",
  speedcast: "#c084fc"
};

const state = {
  data: null,
  activeOperator: "all",
  layers: {
    coverage: L.layerGroup(),
    pops: L.layerGroup(),
    orbits: L.layerGroup()
  },
  satelliteMarkers: [],
  orbitLines: []
};

const map = L.map("map", {
  worldCopyJump: true,
  zoomControl: false
}).setView([23, 0], 2);

L.control.zoom({ position: "bottomleft" }).addTo(map);
L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
  maxZoom: 8,
  attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
}).addTo(map);

state.layers.coverage.addTo(map);
state.layers.pops.addTo(map);
state.layers.orbits.addTo(map);

init();

async function init() {
  try {
    const response = await fetch("./data/dashboard.json", { cache: "no-store" });
    if (!response.ok) throw new Error(`dashboard data ${response.status}`);
    state.data = await response.json();
    renderDashboard();
    requestAnimationFrame(animateSatellites);
  } catch (err) {
    setSelection("Dataset load failed", { error: err.message });
  }
  if (window.lucide) window.lucide.createIcons();
}

function renderDashboard() {
  renderMetrics();
  renderOperatorList();
  renderCoverage();
  renderInfrastructure();
  renderOrbits();
  bindLayerControls();
  const stamp = new Date(state.data.generated_at);
  document.getElementById("generated-at").textContent =
    `Dataset generated ${stamp.toISOString()} · ${state.data.stats.total_prefixes.toLocaleString()} prefixes`;
}

function renderMetrics() {
  const { stats, operators } = state.data;
  text("metric-prefixes", stats.total_prefixes);
  text("metric-announced", stats.announced_prefixes);
  text("metric-pops", stats.prefixes_with_pop);
  text("metric-operators", operators.length);
}

function renderOperatorList() {
  const root = document.getElementById("operator-list");
  root.textContent = "";
  root.appendChild(operatorButton({ name: "all", count: state.data.stats.total_prefixes, orbit_class: "all" }));
  state.data.operators.forEach((op) => root.appendChild(operatorButton(op)));
}

function operatorButton(op) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = `operator-button ${op.name === state.activeOperator ? "active" : ""}`;
  button.dataset.operator = op.name;
  button.title = op.orbit_class;

  const swatch = document.createElement("span");
  swatch.className = "swatch";
  swatch.style.background = op.name === "all" ? "#cbd5e1" : colorFor(op.name);

  const name = document.createElement("span");
  name.className = "name";
  name.textContent = op.name;

  const count = document.createElement("span");
  count.className = "count";
  count.textContent = Number(op.count).toLocaleString();

  button.append(swatch, name, count);
  button.addEventListener("click", () => {
    state.activeOperator = op.name;
    document.querySelectorAll(".operator-button").forEach((el) => {
      el.classList.toggle("active", el.dataset.operator === op.name);
    });
    renderCoverage();
    renderInfrastructure();
    renderOrbits();
  });
  return button;
}

function renderCoverage() {
  state.layers.coverage.clearLayers();
  const max = Math.max(...state.data.countries.map((d) => d.prefixes), 1);
  state.data.countries.forEach((country) => {
    const count = operatorCount(country.operators);
    if (count === 0) return;
    const radius = 90000 + Math.sqrt(count / max) * 850000;
    const marker = L.circle([country.lat, country.lon], {
      radius,
      color: "#14532d",
      weight: 1,
      fillColor: "#22c55e",
      fillOpacity: 0.22
    });
    marker.bindPopup(popupTable("GeoIP coverage", {
      country: country.country,
      prefixes: count.toLocaleString(),
      announced: country.announced.toLocaleString(),
      semantics: "operator-declared customer subnet GeoIP"
    }));
    marker.on("click", () => setSelection("GeoIP coverage", {
      country: country.country,
      prefixes: count,
      announced: country.announced,
      semantics: "customer_subnet_geoip_location"
    }));
    marker.addTo(state.layers.coverage);
  });
}

function renderInfrastructure() {
  state.layers.pops.clearLayers();
  state.data.pops.forEach((pop) => {
    const count = operatorCount(pop.operators);
    if (count === 0) return;
    const icon = L.divIcon({
      className: "",
      html: '<span class="infra-marker pop"></span>',
      iconSize: [14, 14],
      iconAnchor: [7, 7]
    });
    const marker = L.marker([pop.lat, pop.lon], { icon });
    marker.bindPopup(popupTable("PoP assignment", {
      code: pop.code,
      iata: pop.iata,
      country: pop.country,
      prefixes: count.toLocaleString(),
      semantics: "subnet_to_pop_assignment"
    }));
    marker.on("click", () => setSelection("PoP assignment", {
      code: pop.code,
      iata: pop.iata,
      country: pop.country,
      prefixes: count,
      semantics: "subnet_to_pop_assignment"
    }));
    marker.addTo(state.layers.pops);
  });

  state.data.gateways.forEach((gateway) => {
    if (!operatorActive(gateway.operator)) return;
    const icon = L.divIcon({
      className: "",
      html: '<span class="infra-marker gateway"></span>',
      iconSize: [12, 12],
      iconAnchor: [6, 6]
    });
    const marker = L.marker([gateway.lat, gateway.lon], { icon });
    marker.bindPopup(popupTable("Gateway reference", {
      operator: gateway.operator,
      country: gateway.country,
      semantics: gateway.semantics
    }));
    marker.on("click", () => setSelection("Gateway reference", {
      operator: gateway.operator,
      country: gateway.country,
      semantics: gateway.semantics
    }));
    marker.addTo(state.layers.pops);
  });
}

function renderOrbits() {
  state.layers.orbits.clearLayers();
  state.satelliteMarkers = [];
  state.orbitLines = [];
  state.data.orbits.forEach((orbit) => {
    if (!operatorActive(orbit.operator)) return;
    const line = L.polyline(orbitPath(orbit, 0), {
      color: orbit.color,
      weight: orbit.orbit_class === "leo" ? 1.8 : 2.4,
      opacity: orbit.orbit_class === "geo" || orbit.orbit_class.includes("geo") ? 0.46 : 0.58,
      dashArray: orbit.orbit_class === "leo" ? "6 9" : ""
    }).addTo(state.layers.orbits);
    const icon = L.divIcon({
      className: "",
      html: `<span class="sat-marker" style="background:${orbit.color}"></span>`,
      iconSize: [13, 13],
      iconAnchor: [6.5, 6.5]
    });
    const marker = L.marker([0, 0], { icon }).addTo(state.layers.orbits);
    marker.bindPopup(popupTable("Approximate satellite track", {
      operator: orbit.operator,
      orbit: orbit.orbit_class,
      altitude_km: Number(orbit.altitude_km).toLocaleString()
    }));
    state.satelliteMarkers.push({ orbit, marker });
    state.orbitLines.push({ orbit, line });
  });
}

function animateSatellites(now) {
  const t = (now || 0) / 1000;
  state.satelliteMarkers.forEach(({ orbit, marker }) => {
    marker.setLatLng(orbitPoint(orbit, t));
  });
  requestAnimationFrame(animateSatellites);
}

function orbitPath(orbit, t) {
  const points = [];
  for (let i = 0; i <= 360; i += 4) {
    points.push(orbitPoint(orbit, t, i));
  }
  return points;
}

function orbitPoint(orbit, t, offset = 0) {
  const speed = orbit.orbit_class === "leo" ? 11 : orbit.orbit_class === "meo" ? 3.4 : 0.18;
  const angle = ((t * speed + orbit.phase + offset) % 360) * Math.PI / 180;
  const inc = orbit.inclination * Math.PI / 180;
  const lat = Math.sin(angle) * Math.sin(inc) * 82;
  let lon = ((angle * 180 / Math.PI * 1.8 + orbit.phase) % 360) - 180;
  if (orbit.orbit_class === "meo") {
    lon = ((angle * 180 / Math.PI + orbit.phase) % 360) - 180;
  }
  if (orbit.orbit_class.includes("geo")) {
    return [0, ((orbit.phase + offset * 0.08) % 360) - 180];
  }
  return [lat, lon];
}

function bindLayerControls() {
  bindLayer("layer-coverage", state.layers.coverage);
  bindLayer("layer-pops", state.layers.pops);
  bindLayer("layer-orbits", state.layers.orbits);
}

function bindLayer(id, layer) {
  const input = document.getElementById(id);
  input.addEventListener("change", () => {
    if (input.checked) layer.addTo(map);
    else map.removeLayer(layer);
  });
}

function setSelection(title, values) {
  const root = document.getElementById("selection");
  const rows = Object.entries(values)
    .map(([key, value]) => `<dt>${escapeHTML(key)}</dt><dd>${escapeHTML(String(value))}</dd>`)
    .join("");
  root.innerHTML = `<strong>${escapeHTML(title)}</strong><dl>${rows}</dl>`;
}

function popupTable(title, values) {
  const rows = Object.entries(values)
    .map(([key, value]) => `<tr><th>${escapeHTML(key)}</th><td>${escapeHTML(String(value))}</td></tr>`)
    .join("");
  return `<strong>${escapeHTML(title)}</strong><table>${rows}</table>`;
}

function operatorCount(counts) {
  if (state.activeOperator === "all") {
    return Object.values(counts || {}).reduce((sum, value) => sum + Number(value || 0), 0);
  }
  return Number((counts || {})[state.activeOperator] || 0);
}

function operatorActive(operator) {
  return state.activeOperator === "all" || state.activeOperator === operator;
}

function colorFor(operator) {
  return COLORS[operator] || "#cbd5e1";
}

function text(id, value) {
  document.getElementById(id).textContent = Number(value).toLocaleString();
}

function escapeHTML(value) {
  return value.replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#039;"
  })[char]);
}
