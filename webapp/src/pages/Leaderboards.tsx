import { Grid, GridItem } from '@chakra-ui/react'
import { LatestPolls, TopCreators, TopTenPolls, TopVoters } from '../components/Top'

const Leaderboards = () => (
  <Grid
    gap={3}
    templateAreas={{
      base: '"creators" "voters" "polls" "latest"',
      sm: '"creators voters" "polls latest"',
      xl: '"creators voters polls latest"',
    }}
    templateColumns={{ base: 'full', sm: '50%', xl: 'auto auto 30% 30%' }}
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
    <GridItem area='latest'>
      <LatestPolls w='full' />
    </GridItem>
  </Grid>
)

export default Leaderboards
