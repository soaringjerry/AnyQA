import yaml from 'js-yaml';
import configData from './config.yaml?raw'; // 利用Vite原生功能以raw文本导入

const config = yaml.load(configData);

export default config;