module.exports = {
  extends: ['./.eslintrc.js'],
  rules: {
    // RULE 1: Components cannot import services directly
    'no-restricted-imports': ['error', {
      paths: [
        {
          name: '@/services',
          message: '❌ Components cannot import services directly. Use stores instead.'
        }
      ],
      patterns: [
        {
          group: ['*/services/*', '../services/*', '../../services/*'],
          message: '❌ ARCHITECTURE VIOLATION: Components must use stores, not services directly.'
        },
        {
          group: ['*/websocket/*', '*/WebSocketService'],
          message: '❌ ARCHITECTURE VIOLATION: Only APIClient should use WebSocketService.'
        }
      ]
    }],
    
    // RULE 2: Enforce import boundaries by file location
    'import/no-restricted-paths': ['error', {
      zones: [
        // Components can only import from stores, types, and other components
        {
          target: './src/components',
          from: './src/services',
          message: '❌ Components cannot import services. Use stores: import { useXStore } from "@/stores/x"'
        },
        {
          target: './src/pages',
          from: './src/services',
          except: ['./logger/LoggerService.ts'],
          message: '❌ Pages cannot import services. Use stores: import { useXStore } from "@/stores/x"'
        },
        
        // Stores cannot import components
        {
          target: './src/stores',
          from: './src/components',
          message: '❌ Stores cannot import components (circular dependency)'
        },
        {
          target: './src/stores',
          from: './src/pages',
          message: '❌ Stores cannot import pages (circular dependency)'
        },
        
        // Services cannot import stores or components
        {
          target: './src/services',
          from: './src/stores',
          message: '❌ Services cannot import stores (wrong direction). Services should be stateless.'
        },
        {
          target: './src/services',
          from: './src/components',
          message: '❌ Services cannot import components (wrong layer)'
        },
        
        // Only APIClient can import WebSocketService
        {
          target: './src/services/!(abstraction)/**',
          from: './src/services/websocket',
          message: '❌ Only APIClient should import WebSocketService. Use APIClient instead.'
        }
      ]
    }],
    
    // RULE 3: Enforce consistent service constructor patterns
    'no-restricted-syntax': ['error',
      {
        selector: 'NewExpression[callee.name=/.*Service$/][arguments.length=0]',
        message: '❌ Services must receive dependencies via constructor (APIClient and Logger)'
      },
      {
        selector: 'NewExpression[callee.name=/.*Service$/] > Identifier[name="WebSocketService"]',
        message: '❌ Services should use APIClient, not WebSocketService directly'
      }
    ],
    
    // RULE 4: Enforce store naming conventions
    'filename-rules/match': ['error', {
      './**/stores/**/*.ts': 'camelCase',
      './**/stores/**/*Store.ts': /^[a-z]+Store\.ts$/
    }],
    
    // RULE 5: Prevent direct JSON-RPC calls outside of APIClient
    'no-restricted-properties': ['error',
      {
        object: 'wsService',
        property: 'sendRPC',
        message: '❌ Use APIClient.call() instead of direct WebSocket.sendRPC()'
      }
    ]
  },
  
  overrides: [
    {
      // Special rules for test files
      files: ['**/*.test.ts', '**/*.spec.ts'],
      rules: {
        'no-restricted-imports': 'off',
        'import/no-restricted-paths': 'off'
      }
    },
    {
      // APIClient is the ONLY file allowed to use WebSocketService
      files: ['src/services/abstraction/APIClient.ts'],
      rules: {
        'no-restricted-imports': ['error', {
          patterns: [
            {
              group: ['*/services/*', '!*/websocket/WebSocketService'],
              message: 'APIClient can only import WebSocketService'
            }
          ]
        }]
      }
    }
  ]
};
