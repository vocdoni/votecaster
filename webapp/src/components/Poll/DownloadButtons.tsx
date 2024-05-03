import { Button } from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { FaDownload } from 'react-icons/fa6'
import { useAuth } from '~components/Auth/useAuth'
import { fetchPollsRemainingVoters, fetchPollsVoters } from '~queries/polls'
import { CsvGenerator } from '~src/generator'
import { downloadFile } from '~util/files'

export type DownloadUsersListButtonProps = {
  electionId: string
  filename: string
  text: string
  queryFn: () => Promise<string[]>
}

export const DownloadUsersListButton = ({ electionId, filename, text, queryFn }: DownloadUsersListButtonProps) => {
  const {
    data: voters,
    refetch,
    isFetching,
  } = useQuery({
    queryKey: [text, electionId],
    queryFn,
    enabled: false,
  })
  const [downloaded, setDownloaded] = useState<string>('')

  useEffect(() => {
    if (voters?.length && downloaded !== JSON.stringify(voters)) {
      const csv = new CsvGenerator(
        ['Username'],
        voters.map((username) => [username]),
        filename
      )
      setDownloaded(JSON.stringify(voters))
      downloadFile(csv.url, csv.filename)
    }
  }, [voters])

  return (
    <Button
      isLoading={isFetching}
      loadingText='Preparing download...'
      onClick={() => refetch()}
      colorScheme='blue'
      size='sm'
      rightIcon={<FaDownload />}
      disabled={isFetching}
    >
      {text}
    </Button>
  )
}

export const DownloadVotersButton = ({ electionId }: { electionId: string }) => {
  const { bfetch } = useAuth()

  return (
    <DownloadUsersListButton
      electionId={electionId}
      filename='voters.csv'
      text='Download voters list'
      queryFn={fetchPollsVoters(bfetch, electionId)}
    />
  )
}

export const DownloadRemainingVotersButton = ({ electionId }: { electionId: string }) => {
  const { bfetch } = useAuth()

  return (
    <DownloadUsersListButton
      electionId={electionId}
      filename='remaining-voters.csv'
      text='Download remaining voters list'
      queryFn={fetchPollsRemainingVoters(bfetch, electionId)}
    />
  )
}
