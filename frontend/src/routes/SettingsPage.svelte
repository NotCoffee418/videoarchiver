<script>
  import SettingView, { SettingType } from "../components/settings/SettingView.svelte";
  import SettingsGroup from "../components/settings/SettingsGroup.svelte";
  import AllowDuplicatesSetting from "../components/settings/AllowDuplicatesSetting.svelte";
  import Expander from "../components/Expander.svelte";
  
</script>

<SettingsGroup title="General">
    <SettingView 
        key="autostart_service"
        label="Autostart Downloader Service"
        description="This will automatically start the downloader service when the computer starts"
        type={SettingType.BOOL} />
    <SettingView 
        key="autoupdate_ytdlp"
        label="Autoupdate yt-dlp"
        description="This will automatically update yt-dlp when a new version is available"
        type={SettingType.BOOL} />
    <AllowDuplicatesSetting 
        key="allow_duplicates"
        label="Allow Duplicate Downloads"
        description="Disables duplicate detection, allowing videos to be re-downloaded even if they already exist in your directories." />
</SettingsGroup>

<SettingsGroup title="SponsorBlock">
    <div class="flex-container">
        <div class="column">
            <SettingView 
                key="sponsorblock_video"
                label="Video"
                description="SponsorBlock will remove the following segments from videos"
                type={SettingType.MULTISELECT}
                options={[
                    { label: "Sponsor", value: "sponsor" },
                    { label: "Intro", value: "intro" },
                    { label: "Outro", value: "outro" },
                    { label: "Selfpromo", value: "selfpromo" },
                    { label: "Interaction", value: "interaction" },
                    { label: "Music Offtopic", value: "music_offtopic" },
                    { label: "Preview", value: "preview" },
                    { label: "Filler", value: "filler" },
                    { label: "Chapter", value: "chapter" }
                ]} />
        </div>
        <div class="column">
            <SettingView 
                key="sponsorblock_audio"
                label="Music"
                description="SponsorBlock will remove the following segments from music"
                type={SettingType.MULTISELECT}
                options={[
                    { label: "Sponsor", value: "sponsor" },
                    { label: "Intro", value: "intro" },
                    { label: "Outro", value: "outro" },
                    { label: "Selfpromo", value: "selfpromo" },
                    { label: "Interaction", value: "interaction" },
                    { label: "Music Offtopic", value: "music_offtopic" },
                    { label: "Preview", value: "preview" },
                    { label: "Filler", value: "filler" },
                    { label: "Chapter", value: "chapter" }
                ]} />
        </div>
      </div>
      
</SettingsGroup>

<SettingsGroup title="Advanced">
    <Expander>
        <SettingView 
            key="autostart_service_demo"
            label="Autostart Downloader Service"
            description="This will automatically start the downloader service when the computer starts"
            type={SettingType.BOOL} />
    </Expander>
</SettingsGroup>
<!-- Hidden div for development -->
<!-- <SettingsGroup title="Example Settings">
    <SettingView 
        key="string_example"
        label="String Example"
        description="This is a description"
        type={SettingType.STRING}
        validationFunction={(value) => {
            return value.length > 2;
        }} />
<SettingView 
    key="bool_example"
    label="Bool Example"
    type={SettingType.BOOL} />
<SettingView 
    key="int_example"
    label="Int Example"
    type={SettingType.INT}
    validationFunction={(value) => {
        return value > 10;
    }} />
<SettingView 
    key="float_example"
    label="Float Example"
    type={SettingType.FLOAT} />
<SettingView 
    key="combobox_example"
    label="Combobox Example"
    type={SettingType.SELECT}
    options={[
        { label: "Option 1", value: "option1" },
        { label: "Option 2", value: "option2" },
        { label: "Option 3", value: "option3" }
    ]}
    validationFunction={(value) => {
        return value === "option1";
    }} />
<SettingView 
    key="multi_select_example"
    label="Multi Select Example"
    type={SettingType.MULTISELECT}
    options={[
        { label: "Option 1", value: "option1_value" },
        { label: "Option 2", value: "option2_value" },
        { label: "Option 3", value: "option3_value" }
    ]}
        validationFunction={(values) => {
            return values.get("option1_value") === true;
        }} />     
</SettingsGroup> -->

<style>
    .flex-container {
      display: flex;
      gap: 1rem;
    }
    
    .column {
      flex: 1;
    }
    
    @media (max-width: 768px) {
      .flex-container {
        flex-direction: column;
      }
    }
</style>