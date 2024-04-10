import { Grid, GridItem } from '@chakra-ui/react'
import { TopCreators, TopTenPolls, TopVoters } from '../components/Top'

export const Leaderboards = () => (
  <Grid
    gap={3}
    templateAreas={{
      base: '"creators" "voters" "polls"',
      sm: '"creators voters" "polls polls"',
      lg: '"creators voters polls"',
    }}
    templateColumns={{ base: 'full', sm: '50%', lg: 'auto auto 50%' }}
    w='full'
  >
    <GridItem area='polls'>
      <TopTenPolls />
    </GridItem>
    <GridItem area='creators'>
      <TopCreators w='full' />
    </GridItem>
    <GridItem area='voters'>
      <TopVoters w='full' />
    </GridItem>
  </Grid>
)
