// Base URL for the API endpoints – assumes same origin
const API_BASE = "/dashboard/v1";

/**
 * Helper function to perform API requests.
 * @param {string} method - HTTP method (GET, POST, PUT, DELETE)
 * @param {string} endpoint - API endpoint (e.g., "/registrations/")
 * @param {object} data - Data to send as JSON (optional)
 * @returns {Promise<object|string>} - Parsed JSON response or plain text.
 */
async function apiRequest(method, endpoint, data = null) {
  const options = { method, headers: {} };
  if (data) {
    options.headers["Content-Type"] = "application/json";
    options.body = JSON.stringify(data);
  }
  try {
    const response = await fetch(API_BASE + endpoint, options);
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`HTTP error ${response.status}: ${errorText}`);
    }
    const contentType = response.headers.get("Content-Type");
    if (contentType && contentType.includes("application/json")) {
      return await response.json();
    } else {
      return await response.text();
    }
  } catch (error) {
    console.error("API Request error:", error);
    throw error;
  }
}

/**
 * Registers a new dashboard configuration.
 */
async function registerDashboard() {
  const input = document.getElementById("dashboardRegisterInput").value.trim();
  if (!input) {
    alert("Please enter a country name or ISO code.");
    return;
  }
  // Create a default configuration with sample features.
  const requestData = {
    country: "",
    isoCode: "",
    features: {
      temperature: true,
      precipitation: true,
      capital: true,
      coordinates: true,
      population: true,
      area: false,
      targetCurrencies: ["EUR", "USD", "SEK"]
    }
  };

  // Use ISO code if input has exactly two characters, otherwise use country name.
  if (input.length === 2) {
    requestData.isoCode = input;
  } else {
    requestData.country = input;
  }

  try {
    const result = await apiRequest("POST", "/registrations/", requestData);
    document.getElementById("registerDashboardResult").innerHTML =
      `Registered with ID: ${result.id}<br/>Last Change: ${result.lastChange}`;
  } catch (error) {
    document.getElementById("registerDashboardResult").innerHTML =
      `Error: ${error.message}`;
  }
}

/**
 * Retrieves a populated dashboard by its configuration ID.
 */
async function getPopulatedDashboard() {
  const id = document.getElementById("populatedDashboardInput").value.trim();
  if (!id) {
    alert("Please enter a configuration ID.");
    return;
  }
  try {
    const result = await apiRequest("GET", `/dashboards/${id}`);
    document.getElementById("populatedDashboardResult").textContent =
      JSON.stringify(result, null, 2);
  } catch (error) {
    document.getElementById("populatedDashboardResult").textContent =
      `Error: ${error.message}`;
  }
}

/**
 * Retrieves a dashboard configuration by its ID.
 */
async function getDashboardConfig() {
  const id = document.getElementById("dashboardConfigViewInput").value.trim();
  if (!id) {
    alert("Please enter a configuration ID.");
    return;
  }
  try {
    const result = await apiRequest("GET", `/registrations/${id}`);
    document.getElementById("dashboardConfigResult").textContent =
      JSON.stringify(result, null, 2);
  } catch (error) {
    document.getElementById("dashboardConfigResult").textContent =
      `Error: ${error.message}`;
  }
}

/**
 * Lists all stored dashboard configurations.
 */
async function listDashboardConfigs() {
  try {
    const result = await apiRequest("GET", "/registrations/");
    document.getElementById("dashboardConfigsList").textContent =
      JSON.stringify(result, null, 2);
  } catch (error) {
    document.getElementById("dashboardConfigsList").textContent =
      `Error: ${error.message}`;
  }
}

/**
 * Helper: Gather features values explicitly from checkboxes.
 */
function getFeaturesFromForm() {
  // Ensure each checkbox is read explicitly to include false values.
  return {
    temperature: document.getElementById("temperatureCheckbox").checked,
    precipitation: document.getElementById("precipitationCheckbox").checked,
    capital: document.getElementById("capitalCheckbox").checked,
    coordinates: document.getElementById("coordinatesCheckbox").checked,
    population: document.getElementById("populationCheckbox").checked,
    area: document.getElementById("areaCheckbox").checked,
    // For target currencies, we expect the user to input a valid JSON string.
    targetCurrencies: JSON.parse(
      document.getElementById("targetCurrenciesInput").value.trim() || "[]"
    )
  };
}

/**
 * Edits an existing dashboard configuration by its ID.
 * This version sends all fields provided.
 */
async function editDashboardConfig() {
  const id = document.getElementById("dashboardEditInput").value.trim();
  if (!id) {
    alert("Please enter a configuration ID.");
    return;
  }

  // Grab top-level inputs.
  const countryVal = document.getElementById("dashboardEditCountry").value.trim();
  const isoVal = document.getElementById("dashboardEditISO").value.trim();
  const currencyVal = document.getElementById("dashboardEditCurrency").value.trim();
  const features = getFeaturesFromForm(); // This always includes all keys with true/false

  // Build the payload. Only include keys if non-empty.
  const requestData = {};
  if (countryVal) {
    requestData.country = countryVal;
  }
  if (isoVal) {
    requestData.isoCode = isoVal;
  }
  if (currencyVal) {
    requestData.currency = currencyVal;
  }
  // Always include features from the form, since they now have explicit booleans.
  requestData.features = features;

  // If nothing is provided, alert.
  if (Object.keys(requestData).length === 0) {
    alert("No fields to update.");
    return;
  }

  try {
    await apiRequest("PUT", `/registrations/${id}`, requestData);
    document.getElementById("dashboardEditResult").innerHTML = "Configuration updated.";
  } catch (error) {
    document.getElementById("dashboardEditResult").innerHTML = `Error: ${error.message}`;
  }
}

/**
 * Deletes a dashboard configuration by its ID.
 */
async function deleteDashboardConfig() {
  const id = document.getElementById("dashboardDeleteInput").value.trim();
  if (!id) {
    alert("Please enter a configuration ID.");
    return;
  }
  try {
    await apiRequest("DELETE", `/registrations/${id}`);
    document.getElementById("dashboardDeleteResult").innerHTML = "Configuration deleted.";
  } catch (error) {
    document.getElementById("dashboardDeleteResult").innerHTML = `Error: ${error.message}`;
  }
}

/**
 * Registers a new webhook.
 */
async function registerWebhook() {
  const url = document.getElementById("webhookRegisterInput").value.trim();
  if (!url) {
    alert("Please enter a webhook URL.");
    return;
  }
  const country = document.getElementById("webhookRegisterCountry").value.trim();
  const event = document.getElementById("webhookRegisterEvent").value;
  const requestData = {
    url: url,
    country: country,
    event: event
  };
  try {
    const result = await apiRequest("POST", "/notifications/", requestData);
    document.getElementById("webhookRegisterResult").innerHTML =
      `Webhook registered with ID: ${result.id}`;
  } catch (error) {
    document.getElementById("webhookRegisterResult").innerHTML =
      `Error: ${error.message}`;
  }
}

/**
 * Retrieves a webhook by its ID.
 */
async function getWebhook() {
  const id = document.getElementById("webhookViewInput").value.trim();
  if (!id) {
    alert("Please enter a webhook ID.");
    return;
  }
  try {
    const result = await apiRequest("GET", `/notifications/${id}`);
    document.getElementById("webhookViewResult").textContent = JSON.stringify(result, null, 2);
  } catch (error) {
    document.getElementById("webhookViewResult").textContent = `Error: ${error.message}`;
  }
}

/**
 * Lists all registered webhooks.
 */
async function listWebhooks() {
  try {
    const result = await apiRequest("GET", "/notifications/");
    document.getElementById("webhookListResult").textContent = JSON.stringify(result, null, 2);
  } catch (error) {
    document.getElementById("webhookListResult").textContent = `Error: ${error.message}`;
  }
}

/**
 * Deletes a webhook by its ID.
 */
async function deleteWebhook() {
  const id = document.getElementById("webhookDeleteInput").value.trim();
  if (!id) {
    alert("Please enter a webhook ID.");
    return;
  }
  try {
    await apiRequest("DELETE", `/notifications/${id}`);
    document.getElementById("webhookDeleteResult").innerHTML = "Webhook deleted.";
  } catch (error) {
    document.getElementById("webhookDeleteResult").innerHTML = `Error: ${error.message}`;
  }
}

/**
 * Checks the status of the APIs and services.
 */
async function checkStatus() {
  try {
    const result = await apiRequest("GET", "/status/");
    document.getElementById("statusOutput").textContent = JSON.stringify(result, null, 2);
  } catch (error) {
    document.getElementById("statusOutput").textContent = `Error: ${error.message}`;
  }
}