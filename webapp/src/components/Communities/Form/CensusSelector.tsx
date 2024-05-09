import { Box, FormLabel, Text } from '@chakra-ui/react'
import CensusTypeSelector from '~components/CensusTypeSelector'

export const CensusSelector = () => (
  <Box gap={4} display='flex' flexDir='column'>
    <FormLabel htmlFor='census-type'>Set up a default census</FormLabel>
    <Text>
      This census will be set as your default. You have the flexibility to change it at any time and create new ones in
      the future. A snapshot of eligible voters will be made every time you create a new poll.
    </Text>
    <CensusTypeSelector />
  </Box>
)
