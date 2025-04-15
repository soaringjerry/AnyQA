export default {
    message: {
      hello: 'Hello World',
      welcome: 'Welcome to our site!',
      loading: 'Loading configuration...',
      submitting: 'Submitting...',
      success: 'Sent successfully!',
      submitError: 'Submission failed, please try again later',
      configError: 'System initialization failed, please refresh the page',
      placeholder: 'Type here...',
      success: 'Sent successfully!',
      sessionCreated: 'Session created successfully!',
      sessionIdGenerated: 'Generated Session ID:',
      useFollowingLinks: 'Please use the following links:',
      questionPage: 'Question Page:',
      adminPanel: 'Admin Panel:',
      displayScreen: 'Display Screen:',
      scanQrCode: 'Scan the QR code below to open the question page (Index):',
      clickToGenerate: 'Click the button below to generate a new Session ID and get access links.'
    },
    nav: {
      Questionshere: '✨ Questions here ✨',
      about: 'About Us'
    },
    button: {
      send: 'Send Question',
      chinese: '中文',
      english: 'English',
      japanese: '日本語',
      russian: 'Русский',
      spanish: 'Español',
      malay: 'Bahasa Melayu',
      tamil: 'தமிழ்',
      start: 'Start',
      regenerate: 'Regenerate'
    },
    title: {
      createSession: 'Create New Session',
      sessionSetup: 'Quickly generate a new session and get corresponding access links and QR code here.'
    },
    presenter: {
      title: 'Presenter Console (Click question for details)',
      questionDetail: 'Question Details',
      questionContent: 'Question Content',
      aiSuggestion: 'AI Suggestion',
      markFinished: 'Mark as Finished',
      delete: 'Delete',
      show: 'Show Question',
      close: 'Close',
      confirmDelete: 'Are you sure you want to delete this question?',
      noAiSuggestion: 'No AI suggestion available',
      kbSuggestion: 'Knowledge Base Suggestion', // New
      noKbSuggestion: 'No suggestion from knowledge base', // New
      loadError: 'Failed to load questions',
      deleteError: 'Delete failed',
      updateError: 'Update failed',
      showError: 'Failed to show question',
      sessionIdError: 'Missing sessionId parameter',
      // New translations for document management
      uploadTitle: 'Upload Knowledge Base Documents',
      uploadButton: 'Upload Document',
      uploadedDocsTitle: 'Uploaded Documents',
      loadingDocs: 'Loading document list...',
      uploadedAt: 'Uploaded at',
      noDocs: 'No documents uploaded for this session yet.',
      selectFileError: 'Please select a file to upload.',
      noSessionIdError: 'Configuration not loaded or missing session ID, cannot upload.',
      uploading: 'Uploading...',
      uploadSuccess: 'Document "{title}" uploaded successfully! Processing in background...',
      uploadFailed: 'Upload failed',
      uploadError: 'Upload failed: {message}',
      confirmDeleteDoc: 'Are you sure you want to delete the document "{title}"? This will also delete all associated data.',
      deletingDoc: 'Deleting document "{title}"...',
      deleteDocSuccess: 'Document "{title}" deleted successfully!',
      deleteDocFailed: 'Delete failed.',
      deleteDocError: 'Failed to delete document "{title}": {message}',
      loadDocsError: 'Failed to load document list: {message}',
      // New translations for prompt editing
      promptSettingsTitle: 'Prompt Settings',
      genericPromptLabel: 'Generic AI Suggestion Prompt:',
      genericPromptPlaceholder: '(Leave empty to use system default)',
      kbPromptLabel: 'Knowledge Base QA Prompt:',
      kbPromptPlaceholder: '(Leave empty to use system default, must include "%s" for context)',
      kbPromptHint: 'Hint: The knowledge base prompt template must include "%s", which will be replaced by the retrieved document snippets.',
      savePromptsButton: 'Save Prompts',
      loadingPrompts: 'Loading prompts...',
      savingPrompts: 'Saving prompts...',
      loadPromptsError: 'Failed to load prompts: {message}',
      savePromptsSuccess: 'Prompts saved successfully!',
      savePromptsFailed: 'Save failed.',
      savePromptsError: 'Failed to save prompts: {message}'
    },
    display: {
      title: 'Question Display',
      waiting: 'Waiting',
      loadError: 'Failed to load questions',
      initError: 'Initialization failed',
      wsError: 'WebSocket connection failed'
    }
};
  