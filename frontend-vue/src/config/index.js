import yaml from 'js-yaml';
import configData from './config.yaml'; // Temporarily remove ?raw for testing resolution

const config = yaml.load(configData);

export default config;