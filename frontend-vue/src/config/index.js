// Store the loaded config globally within this module
let loadedConfig = null;
let configPromise = null;

// Function to fetch and store the configuration
async function fetchConfig() {
  try {
    // Fetch from the root path served by Nginx. Add cache busting.
    const response = await fetch(`/config.json?t=${new Date().getTime()}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    loadedConfig = await response.json();
    console.log('Configuration loaded:', loadedConfig);
    return loadedConfig;
  } catch (error) {
    console.error('Failed to load configuration:', error);
    // Provide default/fallback configuration or re-throw the error
    // For simplicity, we'll re-throw, forcing the app init to fail clearly
    throw new Error('Could not load application configuration.');
  }
}

// Function to get the configuration, initiates fetch on first call
export function getConfig() {
  if (!configPromise) {
    configPromise = fetchConfig();
  }
  return configPromise;
}

// Export the promise-based getter.
// Components/services can import getConfig and await it.