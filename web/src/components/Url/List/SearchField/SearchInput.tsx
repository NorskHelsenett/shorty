interface SearchInputProps {
  search: string;
  onSearchChange: (value: string) => void;
}

export const SearchInput: React.FC<SearchInputProps> = ({ search, onSearchChange }) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onSearchChange(event.target.value);
  };

  return <input id="search" type="text" value={search} placeholder="Search by Path" onChange={handleChange} className="search" />;
};
