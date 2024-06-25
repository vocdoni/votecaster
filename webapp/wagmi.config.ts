import { defineConfig } from '@wagmi/cli'
import { foundry } from '@wagmi/cli/plugins'

export default defineConfig({
  out: 'src/bindings.ts',
  plugins: [
    foundry({
      project: '../communities',
      exclude: [
        'I*',
        'CommunityHub.t.sol/**',
        'Counter**',
        'Mock*',
        'Script.sol/**',
        'StdError.sol/**',
        'StdStorage.sol/**',
        'StdAssertions.sol/**',
        'StdInvariant.sol/**',
        'Test**',
        'Ownable**',
        'Vm.sol/**',
      ],
    }),
  ],
})
