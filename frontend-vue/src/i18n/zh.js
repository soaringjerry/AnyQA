export default {
    message: {
      hello: '你好，世界',
      welcome: '欢迎来到我们的网站！',
      loading: '正在加载配置...',
      submitting: '提交中...',
      success: '发送成功！',
      submitError: '提交失败，请稍后重试',
      configError: '系统初始化失败，请刷新页面重试',
      placeholder: '在这里输入...',
      success: '发送成功！',
      sessionCreated: '会话已成功生成！',
      sessionIdGenerated: '已生成的 Session ID:',
      useFollowingLinks: '请使用以下链接访问：',
      questionPage: '提问页面：',
      adminPanel: '管理后台：',
      displayScreen: '展示大屏：',
      scanQrCode: '扫描下方二维码直接打开提问页面 (Index)：',
      clickToGenerate: '点击下方按钮生成一个新的 Session ID，并获得相应的访问链接。'
    },
    nav: {
      Questionshere: '✨ 提问空间 ✨',
      about: '关于我们'
    },
    button: {
      send: '发送问题',
      chinese: '中文',
      english: 'English',
      japanese: '日本語',
      russian: 'Русский',
      spanish: 'Español',
      malay: 'Bahasa Melayu',
      tamil: 'தமிழ்',
      start: '开始',
      regenerate: '重新生成'
    },
    title: {
      createSession: '创建新的会话',
      sessionSetup: '在这里快速生成一个全新会话并获取对应的访问链接和二维码。'
    },
    presenter: {
      title: '演讲者控制台（点击问题查看详情页）',
      questionDetail: '问题详情',
      questionContent: '问题内容',
      aiSuggestion: 'AI 建议回复',
      markFinished: '标记完成',
      delete: '删除',
      show: '显示问题',
      close: '关闭',
      confirmDelete: '确定要删除这个问题吗？',
      noAiSuggestion: '暂无AI建议',
      kbSuggestion: '知识库建议', // 新增
      noKbSuggestion: '暂无知识库建议', // 新增
      loadError: '加载问题失败',
      deleteError: '删除失败',
      updateError: '更新失败',
      showError: '设置显示问题失败',
      sessionIdError: '缺少 sessionId 参数',
      // 新增文档管理相关翻译
      uploadTitle: '上传知识库文档',
      uploadButton: '上传文档',
      uploadedDocsTitle: '已上传文档',
      loadingDocs: '正在加载文档列表...',
      uploadedAt: '上传于',
      noDocs: '当前会话没有已上传的文档。',
      selectFileError: '请选择要上传的文件。',
      noSessionIdError: '配置未加载或缺少会话ID，无法上传。',
      uploading: '正在上传...',
      uploadSuccess: '文档 "{title}" 上传成功！后台正在处理...',
      uploadFailed: '上传失败',
      uploadError: '上传失败: {message}',
      confirmDeleteDoc: '确定要删除文档 "{title}" 吗？这将同时删除所有相关数据。',
      deletingDoc: '正在删除文档 "{title}"...',
      deleteDocSuccess: '文档 "{title}" 删除成功！',
      deleteDocFailed: '删除失败。',
      deleteDocError: '删除文档 "{title}" 失败: {message}',
      loadDocsError: '加载文档列表失败: {message}',
      // 新增提示词编辑相关翻译
      promptSettingsTitle: '提示词设置',
      genericPromptLabel: '通用 AI 建议提示词:',
      genericPromptPlaceholder: '（留空则使用系统默认）',
      kbPromptLabel: '知识库问答提示词:',
      kbPromptPlaceholder: '（留空则使用系统默认，必须包含 "%s" 以插入文档内容）',
      kbPromptHint: '提示：知识库提示词模板中必须包含 "%s" 占位符，它将被实际检索到的文档片段替换。',
      savePromptsButton: '保存提示词',
      loadingPrompts: '正在加载提示词...',
      savingPrompts: '正在保存提示词...',
      loadPromptsError: '加载提示词失败: {message}',
      savePromptsSuccess: '提示词保存成功！',
      savePromptsFailed: '保存失败。',
      savePromptsError: '保存提示词失败: {message}'
    },
    display: {
      title: '问题展示',
      waiting: '等待中',
      loadError: '加载问题失败',
      initError: '初始化失败',
      wsError: 'WebSocket 连接失败'
    }
  };
  